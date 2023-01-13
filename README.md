
# Blockchain-explorer

The Blockchain explorer engine component is intended to synchronize the database with the blockchain. Program can be run in manual or automatic mode. Manual mode will perform one synchronization process to the latest block on the blockchain at that moment, while automatic mode monitors the appearance of a new block on the blockchain and trigger the synchronization process upon arrival of the notification.

## Configurations

Use command line arguments to override the default values from the .env file.

Options:
- `--checkpoint.window value` uint <br>
        Sets after how many created blocks the checkpoint is determined
- `--db.host` string <br>
        Database server host
- `--db.name` string <br>
        Database name
- `--db.password` string <br>
        Database user password
- `--db.port` string <br>
        Database server port
- `--db.ssl` string <br>
        Enable (verify-full) or disable TLS
- `--db.user` string <br>
        Database user
- `--http.addr` string <br>
        Blockchain node HTTP address
- `--mode` string <br>
        Manual or automatic mode of application
- `--step value` uint <br>
        Number of requests in one batch sent to the blockchain
- `--timeout value` uint <br>
        Sets a timeout used for requests sent to the blockchain
- `--workers value` uint <br>
        Number of goroutines to use for fetching data from blockchain
- `--ws.addr` string <br>
        Blockchain node WebSocket address