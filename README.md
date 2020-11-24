# go-harvest

go-harvest is a Go client library for accessing the [Harvest API](https://help.getharvest.com/api-v2/).

## Usage

```go
client := harvest.NewClient("accessToken", "accountId")

// list all time entries
timeEntries, _, err := client.TimeEntries.List(ctx, nil)
```

## License

This library is distributed under the BSD-style license found in the
[LICENSE](./LICENSE) file.
