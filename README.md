Betarigs Autorent
=================

This tool will help you to rent hash power on Betarigs.

![Demo](http://dl.toorop.fr/pics/brAutorent-demo.gif)


## Getting Started

Grab the latest release binary:

* [Windows](http://dl.toorop.fr/softs/brAutorent/windows/brAutorent.exe)
* [MacOS](http://dl.toorop.fr/softs/brAutorent/macos/brAutorent)
* [Linux](http://dl.toorop.fr/softs/brAutorent/linux/brAutorent)

If you prefer to build from source:

First [Install Go ](http://golang.org/doc/install)

And run the following commands:

	go get github.com/Toorop/betarigs-autorent
	go build -o brAutorent github.com/Toorop/betarigs-autorent
	
On windows you should replace "-o brAutorent" by "-o brAutorent.exe" 	
## Usage
	$ ./brAutorent --help
	NAME:
   		brAutorent - Rent rigs on Betarig

	USAGE:
   		brAutorent --algo value --mhs value --duration value --maxprice value

   	Example, to rent 10 Mh/s of power for X11 mining during 3 hours at 0.0004 BTC/Mhs/Day to mine on pool "stratum2.suchpool.pw:3335" with worker name "Toorop.Miner1" and worker password "x":
   
   	brAutorent --algo x11 --mhs 10 --duration 3 --maxprice 0.0004	--poolurl stratum2.suchpool.pw:3335 --wname Toorop.Miner1 --wpassword x
   	
   	Before using brAutorent you need to add, in the same folder as this app, a text file named keyring.txt whith:
   	
   	On the first line: 	    Your coinbase API key
   	On the second line: 	Your coinbase API secret
   	On the third line : 	Your betarigs API key

      Rig blacklist:
      If there some rigs that you doesn't want to rent just add a text file named "blacklistedRigs.txt" in the same folder as the app and put ids of the rigs you want to blacklist, ONE PER LINE.
      Lines beginning by # are traited as comment.


	OPTIONS:
   	--algo 		Mining algorithm (scrypt or x11 or x13 or x15 or sha256 or blake256 or scrypt-n or keccak). Required.
   	--mhs '0'		Max mining power to rent in Mh/s. Float. Required.
   	--duration '0'	A given mining duration in hours. Integer. Required.
   	--maxprice '0'	The maximum price in BTC/Mh/day of the rigs to rent. Float. Required.
   	--poolurl 		The pool url formated using host:port, example stratum1.suchpool.pw:3335. Required.
   	--wname 		The pool worker name. Required.
   	--wpassword 		The pool worker password. Required.
   	--dryrun		If this flag is set, brAutorent will simulate the rental creation and payment.
   	--version, -v	print the version
   	--help, -h		show help	 



### Donate

I've worked hard to make this tool useful and easy to use. I've also
released it with a liberal open source license, so that you can do
with it as you please. So, if you find it helpful, I encourage you to
donate what you believe would have been a fair price :

BTC Address: 1FnB6S6TC5Z9T7AT4QkPCZRUrBPBmZHoUA

![Donation QR](http://dl.toorop.fr/pics/btc-address-github.png)