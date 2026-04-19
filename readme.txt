Architecture:
Separation of concern- each logic part was separated into different sub-module (parser, geo mapping, store, report)

Implementing Interface architecture to enable swapping implementation in the future:
	- Can change parsing logic without changing the program structure to support different files formats, or extracting different data to support different dimensions.
	- The in-memory data store can be swapped with external DB (like MySQL, or Redis for cache) to support growing amount of data in the future. I used in memory to simplify the implementation for the exercise 

The conversion from IP to Country was implemented as part of the parsing process to avoid going through the entries again for converting. Doing it in one-go. 

Using locks to keep the code race-condition safe

Assumptions: 
	- The log file is saved in a known path 
	- Log line which failed to be parsed are being reported and we continue 	processing all other lines. 
	- Overlooking browsers and OS versions (like in the example)

How to run:
go run cmd/main.go -file "path/to/your/log.txt" -mmdb "path/to/your/GeoLite2-Country.mmdb"
