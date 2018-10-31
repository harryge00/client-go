package main

import (
	"database/sql"
	"fmt"
	"strings"
	_ "github.com/go-sql-driver/mysql"
	"flag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"bytes"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

const URL = "http://10.30.100.10:8090/api/net/ip/release"

type IpRelease struct {
	IP     string `json:"ip,omitempty"`
	Group  string `json:"group,omitempty"`
	UserId int    `json:"userId,omitempty"`
}

type IpGroupClear struct {
	Group  string `json:"group,omitempty"`
	UserId int    `json:"userId,omitempty"`
}

type IpReleaseResp struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

var IPUser struct {
	IP string
	UserID string
}
func main() {
	db, err := sql.Open("mysql", "root:123456@tcp(10.30.100.6:3306)/net_manager")
	if err != nil {
		panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// Execute the query
	rows, err := db.Query("select * from alloc_ip where occupied=1;")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		// Now do something with the data.
		// Here we just print each column as a string.
		var value, userID, ip, group string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			//fmt.Println(columns[i], ": ", value)
			if columns[i] == "user_id" {
				userID = value
			} else if columns[i] == "ip" {
				ip = value
			} else if columns[i] == "grp" {
				group = value
			}
		}
		uid, _ := strconv.Atoi(userID)
		req := IpRelease{
			IP:     ip,
			UserId: uid,
			Group: group,
		}
		fmt.Println(req)
		//bytes, err := json.Marshal(req)
		//if err != nil {
		//	fmt.Errorf("json errorf %v", err)
		//	continue
		//}
		//_, err = sendReleaseIpReq(bytes, "http://10.30.100.10:8090/api/net/ip/release")
		//if err != nil {
		//	fmt.Println(err)
		//}
	}

}

func sendReleaseIpReq(reqBytes []byte, url string) (code int, err error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var ipResp IpReleaseResp
	err = json.Unmarshal(body, &ipResp)
	if err != nil {
		return
	}
	code = ipResp.Code
	if code != 200 {
		err = fmt.Errorf("%v", ipResp.Message)
	}
	return
}

func releaseIP(userId, ip string) error {

	// Currently, no userid 0
	if userId == "0" {
		return nil
	}
	id, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	req := IpRelease{
		IP:     ip,
		UserId: id,
	}

	bytes, err := json.Marshal(req)
	if err != nil {
		fmt.Errorf("json errorf %v", err)
	}
	_, err = sendReleaseIpReq(bytes, URL)
	fmt.Printf("releaseIP %v %v\n", req, err)
	return err

	return nil
}