
# Tool to help produce statistics from logs

A small log parser made mainly to parse the logs for our server. Only configured for PIF for now. Can be worked on to read other logs.



## How to use

- Have a folder called logs
- Make sure it has all the log files you want to parse
- The logs should be named something like `server_2025-06-30_12-33-26.log`
- The name is the year-month-date-time in 24h format
- All the log files would be fetched from the clients deployed pc
- Run the .exe
- It will make two files. A JSON and a .txt with the name usage_stats
- We open the usage_stats.txt and copy the contents or directly share it


## How it works

It goes through all the files in the folder called logs. Then looks specifically for POST requests made in them as it signifies an attempt to do something on the server. What it looks like in the log file:
```
[GIN] 2025/07/09 - 16:56:31 | 500 |    107.0592ms |             ::1 | POST     "/modes"
```
Right now it looks for the following:

- Modes
- Shades
- IPTV
- AC
- Cyviz

You can find them aliased in the code like:

```
"/modes":         "Room Mode Changes",
"/lutron/shades": "Shade Controls",
"/iptv/channel":  "TV Controls",
"/iptv/remote":   "TV Controls",
"/iptv":          "TV Controls",
"/bacnet/info":   "AC Temperature",
"/cyviz/avinput": "Cyviz TV Controls",
```



## A brief example Log File

```
SERVER USAGE STATISTICS REPORT
===============================

Generated: 2025-01-01 10:00:00

OVERVIEW
--------
Total POST Requests: 500

ENDPOINT USAGE
--------------
Shade Controls      :   100 requests (10.0%)
TV Controls         :   200 requests (20.0%)
Room Mode Changes   :    50 requests (5.0%)
Cyviz TV Controls   :   100 requests (10.0%)
AC Temperature      :    50 requests (5.0%)
...
```

This is just a small example, the file produced will have more info. It also makes a JSON which we can use in other programs to query the data. And **the one it produces in the txt file is what we can share with other**

## Run Locally

Clone the project

```bash
  git clone https://github.com/Shariq-Mersvo/LogParser
```

Go to the project directory

```bash
  cd LogParser
```

Install dependencies

```bash
  go mod tidy
```

Start the server

```bash
  go run main.go
```

## Contributing

Contributions are always welcome to further improve this to be used for all our projects. You can either fork it and work on it or make a pull request to be merged or let me know and I can add you to the project as a contributor

