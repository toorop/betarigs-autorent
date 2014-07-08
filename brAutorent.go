package main

import (
	"bufio"
	"fmt"
	"github.com/Toorop/go-betarigs"
	"github.com/Toorop/go-coinbase"
	"github.com/codegangsta/cli"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	version = "0.0.1"
	fees
)

// Globals
var (
	basePath string
	keyring  *Keyring
	Betarigs *betarigs.Betarigs
	Coinbase *coinbase.Coinbase
	avAlgos  map[string]int
	dryrun   bool
)

type Keyring struct {
	cbApiKey    string
	cbApiSecret string
	brApiKey    string
}

// Load loads keyring from text file
func (k *Keyring) Load() {
	// Open keyring.txt file
	filepath := basePath + "/keyring.txt"
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalln("Unable to open keyring. Be sure you have a kering.txt file with write access located at", filepath, err)
	}
	defer file.Close()

	r := bufio.NewReader(file)
	cbApiKey, _, err := r.ReadLine()
	if err != nil {
		log.Fatalln("Unable to load Coinbase API key from", filepath, ". Type brAutorent --help for info about this file.")
	}
	k.cbApiKey = string(cbApiKey)

	cbApiSecret, _, err := r.ReadLine()
	if err != nil {
		log.Fatalln("Unable to load Coinbase API secret from", filepath, ". Type brAutorent --help for info about this file.")
	}
	k.cbApiSecret = string(cbApiSecret)

	brApiKey, _, err := r.ReadLine()
	if err != nil {
		log.Fatalln("Unable to load Betarigs API key from", filepath, ". Type brAutorent --help for info about this file.")
	}
	k.brApiKey = string(brApiKey)
}

// dieError exit displaying an error message
func dieError(msg ...interface{}) {
	m := "ERROR: "
	for _, v := range msg {
		m += fmt.Sprintf("%v ", v)
	}
	log.Fatalln(m)
}

// dieOk exit
func dieOk(msg ...interface{}) {
	out(msg...)
	os.Exit(0)
}

func out(msg ...interface{}) {
	m := ""
	for _, v := range msg {
		m += fmt.Sprintf("%v ", v)
	}
	if len(m) != 0 {
		log.Println(m)
	}
}

// isValidAlgo check if a given algo exists
func isValidAlgo(algo string) bool {
	for a, _ := range avAlgos {
		if a == algo {
			return true
		}
	}
	return false
}

// getSpeedInMhs returns rigs speed in Mhs/s
func getSpeedInMhs(rig *betarigs.Rig) float64 {
	switch rig.DeclaredSpeed.Unit {
	case "Kh/s":
		return float64(rig.DeclaredSpeed.Value) / 1000
	case "Mh/s":
		return float64(rig.DeclaredSpeed.Value)
	case "Th/s":
		return float64(rig.DeclaredSpeed.Value) * 1000
	default:
		dieError("Unexpected unit value for hashing speed unit:", rig.DeclaredSpeed.Unit)
	}
	return 0.0 // should never happen
}

// getPriceInBtcMhDay returns price in BTC/Mhs/Day
func getPriceInBtcMhDay(rig *betarigs.Rig) float64 {
	switch rig.Price.PerSpeedUnit.Unit {
	case "BTC/Mh/day":
		return float64(rig.Price.PerSpeedUnit.Value)
	case "BTC/Th/day":
		return float64(rig.Price.PerSpeedUnit.Value) / 1000
	default:
		dieError("Unexpected unit value for hashing price unit:", rig.Price.PerSpeedUnit.Unit)
	}
	return 1000000.0 // 1 million BTC per Mh/day ! Amazing ! (don't worry this will never happened)
}

// durationIsAvailable return true duration is available false else
func durationIsAvailable(rig *betarigs.Rig, duration int) bool {
	for _, d := range rig.RentalDurations {
		if d.Unit == "hour" && d.Value == duration {
			return true
		}
	}
	return false
}

// findMatchingRigs return rigs that matche search
func findMatchingRigs(algo string, duration int, mhs, maxprice float64) (rigs []betarigs.Rig, totalSpeed, totalPrice float64, err error) {
	page := 0
	totalSpeed = 0.00
	totalPrice = 0.00
L:
	for {
		page++
		avRigs, err := Betarigs.GetRigs(uint32(avAlgos[algo]), "available", uint32(page))
		if err != nil {
			break L
		}
		if len(avRigs) == 0 {
			break L
		}

		for _, r := range avRigs {
			// Duration
			if !durationIsAvailable(&r, duration) {
				continue
			}

			// price
			price := getPriceInBtcMhDay(&r)
			if price > maxprice {
				break L
			}

			// speed
			speed := getSpeedInMhs(&r)
			if totalSpeed+speed > mhs {
				continue
			}

			// add
			rigs = append(rigs, r)
			totalSpeed += speed
			totalPrice += ((price * speed) / 24) * float64(duration)
		}
	}
	return
}

// Rentrig rents the rig "rig"
func rentRig(rig betarigs.Rig, duration int, pool *betarigs.Pool, chanReply chan bool) {
	if dryrun {
		out("I'm running in dryrun, i will not rent rig", rig.Id)
		chanReply <- true
		return
	}
	// debug
	//rig.Id = 4568

	resp, err := Betarigs.RentRig(rig.Id, duration, pool)
	if err != nil {
		out("ERROR: unable to rent rig", rig.Id, err)
		chanReply <- true
		return
	}
	//out(resp)
	// Rental initialized
	out("New rental", resp.Id, "for rig", rig.Id, "accepted by Betarigs. Rig configuration is done with success.")

	// to avoid Coinbase flooding we will introduce a random sleep (0 to 60 seconds)
	rand.Seed(time.Now().UnixNano())
	tts := rand.Intn(60)
	out("Wait a random delay before paying rental", resp.Id, "to be fair with Coinbase:", tts, "sec")
	time.Sleep(time.Duration(tts) * time.Second)
	toSend := &coinbase.SmTransaction{
		Amount:  strconv.FormatFloat(resp.Payment.Bitcoin.Price.Value, 'f', -1, 64),
		To:      resp.Payment.Bitcoin.PaymentAddress,
		UserFee: "0.0002",
		Idem:    fmt.Sprintf("%d", resp.Id),
	}
	r, err := Coinbase.SendMoney(toSend)
	if err != nil {
		out("ERROR: unable to send", resp.Payment.Bitcoin.Price.Value, "BTC for rental", resp.Id, ".", err)
		chanReply <- true
		return
	}
	out(resp.Payment.Bitcoin.Price.Value, "BTC paid to address", resp.Payment.Bitcoin.PaymentAddress, "for rental", resp.Id)
	// Get tx id and blockchain link
	time.Sleep(1 * time.Second)
	details, err := Coinbase.GetTransactionDetails(r.Transaction.Id)
	if err == nil {
		out("Check transaction processing for rental", resp.Id, fmt.Sprintf("here https://blockchain.info/tx/%s", details.Hsh))
	}
	chanReply <- true
}

func init() {
	var err error
	basePath, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}
	avAlgos = make(map[string]int, 8)
	avAlgos["scrypt"] = 1
	avAlgos["keccak"] = 2
	avAlgos["scrypt-n"] = 3
	avAlgos["sha256"] = 4
	avAlgos["x11"] = 5
	avAlgos["blake256"] = 6
	avAlgos["x13"] = 7
	avAlgos["x15"] = 8
}

func main() {
	app := cli.NewApp()
	app.Name = "brAutorent"
	app.Usage = "Rent rigs on Betarig"
	app.Version = version
	app.Author = "Stéphane Depierrepont aka Toorop"
	app.Email = "toorop@toorop.fr"
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.Name}} --algo value --mhs value --duration value --maxprice value

   Example, to rent 10 Mh/s of power for X11 mining during 3 hours at 0.0004 BTC/Mhs/Day 
   to mine on pool "stratum2.suchpool.pw:3335" with worker name "Toorop.Miner1" and 
   worker password "x":
   {{.Name}} --algo x11 --mhs 10 --duration 3 --maxprice 0.0004	--poolurl stratum2.suchpool.pw:3335 --wname Toorop.Miner1 --wpassword x

   Before using {{.Name}} you need to add, in the same folder as this app, a text file 
   named keyring.txt whith:
   On the first line: 	Your coinbase API key
   On the second line: 	Your coinbase API secret
   On the third line : 	Your betarigs API key


OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}
`

	app.Flags = []cli.Flag{
		cli.StringFlag{"algo", "", "Mining algorithm (scrypt or x11 or x13 or x15 or sha256 or blake256 or scrypt-n or keccak). Required."},
		cli.Float64Flag{"mhs", 0.00, "Max mining power to rent in Mh/s. Float. Required."},
		cli.IntFlag{"duration", 0, "A given mining duration in hours. Integer. Required."},
		cli.Float64Flag{"maxprice", 0.00, "The maximum price in BTC/Mh/day of the rigs to rent. Float. Required."},
		cli.StringFlag{"poolurl", "", "The pool url formated using host:port, example stratum1.suchpool.pw:3335. Required."},
		cli.StringFlag{"wname", "", "The pool worker name. Required."},
		cli.StringFlag{"wpassword", "", "The pool worker password. Required."},
		cli.BoolFlag{"dryrun", `If this flag is set, brAutorent will simulate the rental creation and payment.`},
	}

	app.Action = func(c *cli.Context) {
		// Check inputs
		// Algo
		if !c.IsSet("algo") {
			dieError("Flag --algo is required.")
		}
		algo := strings.ToLower(c.String("algo"))
		if !isValidAlgo(algo) {
			dieError(algo, "is not a valid algorithm.")
		}

		// mhs
		if !c.IsSet("mhs") {
			dieError("Flag --mhs is required.")
		}
		mhs := c.Float64("mhs")
		// duration
		if !c.IsSet("duration") {
			dieError("Flag --duration is required.")
		}
		duration := c.Int("duration")

		// maxprice
		if !c.IsSet("maxprice") {
			dieError("Flag --maxprice is required.")
		}
		maxprice := c.Float64("maxprice")

		// Pool
		pool := new(betarigs.Pool)
		if !c.IsSet("poolurl") {
			dieError("Flag --poolurl is required.")
		}
		pool.Url = c.String("poolurl")

		if !c.IsSet("wname") {
			dieError("Flag --wname is required.")
		}
		pool.WorkerName = c.String("wname")

		if !c.IsSet("wpassword") {
			dieError("Flag --wpassword is required.")
		}
		pool.WorkerPassword = c.String("wpassword")

		// dryrun
		dryrun = c.Bool("dryrun")
		if dryrun {
			out("Running in dry run mode.")
		}

		// populate keyring
		keyring = new(Keyring)
		keyring.Load()

		// Init betarigs
		Betarigs = betarigs.New(string(keyring.brApiKey))

		// Init coinbase
		Coinbase = coinbase.New(keyring.cbApiKey, keyring.cbApiSecret)

		rigs, totalSpeed, totalPrice, err := findMatchingRigs(algo, duration, mhs, maxprice)
		if err != nil {
			dieError("while searching matching rigs:", err)
		}
		if len(rigs) == 0 {
			dieOk("Sorry i've found nothing :(")
		}
		out(fmt.Sprintf("Found %f Mh/s for %d hours renting at the total price of %f BTC (avg %f BTC/Mhs/d)", totalSpeed, duration, totalPrice, ((totalPrice/totalSpeed)/3)*24))

		// Get user BTC balance on coinbase
		// We check his primary account only
		btcBalance, err := Coinbase.GetPrimaryAccountBalance()
		if err != nil {
			dieError("Fail to get your current coinbase account balance", err)
		}

		// Total including 0.002 BTC fees per tx
		totalPrinceIncFees := totalPrice + float64(0.002)*float64(len(rigs)) // hum....
		if btcBalance < totalPrinceIncFees {
			dieError("Hum... i'm sorry to inform you that you don't have enought BTC on your Coinbase primary account. You need", totalPrinceIncFees, " BTC (rent + tx fees) but you have only ", btcBalance, " BTC.")
		}

		// The race begin....
		chanRentReply := make(chan bool)
		for _, rig := range rigs {
			go rentRig(rig, duration, pool, chanRentReply)
		}
		i := 0
		for {
			select {
			case <-chanRentReply:
				out("Renting", fmt.Sprintf("%d/%d", i+1, len(rigs)), "completed.")
			}
			i++
			if i >= len(rigs) {
				break
			}
		}
		dieOk("All my jobs are done. Bye.")
	}

	app.Run(os.Args)
}
