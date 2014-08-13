package main

import (
	td_client "github.com/treasure-data/td-client-go"
	"os"
	"fmt"
	"time"
	"bytes"
	"strconv"
	"compress/gzip"
	"github.com/ugorji/go/codec"
)

func CompressWithGzip(b []byte) []byte {
	retval := bytes.Buffer {}
	w := gzip.NewWriter(&retval)
	w.Write(b)
	w.Close()
	return retval.Bytes()
}

func main() {
	apiKey := os.Getenv("TD_CLIENT_API_KEY")
	client, err := td_client.NewTDClient(td_client.Settings {
		ApiKey: apiKey,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	status, err := client.ServerStatus()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("status: %s\n", status.Status)
	account, err := client.ShowAccount()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("account:")
	fmt.Printf("  id: %d\n", account.Id)
	fmt.Printf("  plan: %d\n", account.Plan)
	fmt.Printf("  storageSize: %d\n", account.StorageSize)
	fmt.Printf("  guaranteedCores: %d\n", account.GuaranteedCores)
	fmt.Printf("  maximumCores: %d\n", account.MaximumCores)
	fmt.Printf("  createdAt: %s\n", account.CreatedAt.Format(time.RFC3339))
	results, err := client.ListResults()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%d results\n", len(*results))
	databases, err := client.ListDatabases()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%d databases\n", len(*databases))
	for _, database := range *databases {
		fmt.Printf("  name: %s\n", database.Name)
		tables, err := client.ListTables(database.Name)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("  %d tables\n", len(*tables))
		for _, table := range *tables {
			fmt.Printf("    name: %s\n", table.Name)
			fmt.Printf("    type: %s\n", table.Type)
			fmt.Printf("    count: %d\n", table.Count)
			fmt.Printf("    primaryKey: %s\n", table.PrimaryKey)
			fmt.Printf("    schema: %v\n", table.Schema)
		}
	}
	err = client.CreateDatabase("sample_db2", nil)
	if err != nil {
		_err := err.(*td_client.APIError)
		if _err == nil || _err.Type != td_client.AlreadyExistsError {
			fmt.Println(err.Error())
			return
		}
	}
	err = client.CreateLogTable("sample_db2", "test")
	if err != nil {
		_err := err.(*td_client.APIError)
		if _err == nil || _err.Type != td_client.AlreadyExistsError {
			fmt.Println(err.Error())
			return
		}
	} else {
		err = client.UpdateSchema("sample_db2", "test", []interface{} {
			[]string { "a", "string" },
			[]string { "b", "string" },
		})
	}
	data := bytes.Buffer {}
	handle := codec.MsgpackHandle {}
	encoder := codec.NewEncoder(&data, &handle)
	for i := 0; i < 10000; i += 1 {
		encoder.Encode(map[string]interface{} {
			"time":i, "a": strconv.Itoa(i), "b": strconv.Itoa(i),
		})
	}
	payload := CompressWithGzip(data.Bytes())
	fmt.Printf("payloadSize:%d\n", len(payload))
	time_, err := client.Import("sample_db2", "test", "msgpack.gz", (td_client.InMemoryBlob)(payload), "")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("elapsed time:%g\n", time_)
	jobId, err := client.SubmitQuery("sample_db2", td_client.Query {
		Type: "hive",
		Query: "SELECT COUNT(*) AS c FROM test WHERE a >= 5000",
		ResultUrl: "",
		Priority: 0,
		RetryLimit: 0,
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("jobId:%s\n", jobId)
	for {
		status, err := client.JobStatus(jobId)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("jobStatus:%s\n", status)
		if status != "queued" && status != "running" {
			break
		}
		time.Sleep(1000000000)
	}
	err = client.JobResultEach(jobId, func(v interface{}) error {
		fmt.Printf("Result:%v\n", v)
		return nil
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	type_, err := client.DeleteTable("sample_db2", "test")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("deleteTable result: %s\n", type_)
	err = client.DeleteDatabase("sample_db2")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
