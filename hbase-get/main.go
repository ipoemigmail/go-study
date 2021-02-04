package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
)

// Quorums is
const Quorums = "localhost:2181"

// Scan is
func Scan(client gohbase.Client, table string, contentParser func([]byte) interface{}) (keys []string, values []interface{}) {
	scan, _ := hrpc.NewScan(context.Background(), []byte(table))
	scanResult := client.Scan(scan)
	keys = make([]string, 0)
	values = make([]interface{}, 0)
	for {
		row, err := scanResult.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}
		keys = append(keys, string(row.Cells[0].Row))
		values = append(values, contentParser(row.Cells[0].Value))
	}
	return
}

func main() {
	client := gohbase.NewClient(Quorums)
	tableName := "table"
	//keys, values := Scan(client, tableName, func(x []byte) interface{} {
	//	var count int64
	//	_ = binary.Read(bytes.NewReader(x), binary.BigEndian, &count)
	//	return count
	//})
	keys, values := Scan(client, tableName, func(x []byte) interface{} {
		return string(x)
	})
	const MaxInt64 = int64(^uint64(0) >> 1)

	//10017607
	//09223370424497420782

	fmt.Printf("%d\n", MaxInt64)
	fmt.Printf("%d\n", MaxInt64-1612357354988)

	for i := range keys {
		if strings.HasPrefix(keys[i], "973_000000000041973") {
			fmt.Printf("%s ==> %s\n", keys[i], values[i])
		}
	}
}
