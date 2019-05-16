# clockifytoxls

A simple golang script (WIP) using [excelize](https://github.com/360EntSecGroup-Skylar/excelize) to read time entries from the clockify api and save them as an xlsx file.

If the clockify time entry description contains a ":", it will be split into two columns, using ":" as seperator.

## Usage

In the project root, create a `config.json` file in the following format to use:

```json
{
    "workspaceId" : "",
    "userId" : "",
    "apiKey" : ""
}
```

Run the script using `go run main.go` or using a custom start date: `go run main.go -s YYYY-MM-DD`.