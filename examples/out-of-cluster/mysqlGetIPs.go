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

	ipToUsers := make(map[string]string)
	ipToGroups := make(map[string]string)
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
		ipToUsers[ip] = userID
		ipToGroups[ip] = group
		//fmt.Println("-----------------------------------")
	}
	fmt.Println(ipToUsers)
	if err = rows.Err(); err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	kubeconfig := flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
	flag.Parse()
	// uses the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	pods, err := clientset.CoreV1().Pods("").List(meta_v1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	//for _, po := range pods.Items {
	//	if po.Annotations["ips"] != "" {
	//		strArr := strings.Split(po.Annotations["ips"], "-")
	//		if len(strArr) == 2 {
	//			ip := strArr[1]
	//			fmt.Println(ip)
	//			if _, exists := ipToUsers[ip]; !exists {
	//				fmt.Printf("%v from K8s does not appear in DB!\n", ip)
	//			} else {
	//				ipToUsers[strArr[1]] = "found"
	//			}
	//		}
	//	}
	//}
	//fmt.Println("IPs occupied according to DB but not in K8s")
	//for ip := range ipToUsers {
	//	if ipToUsers[ip] != "found" {
	//		err = releaseIP(ipToUsers[ip], ip)
	//		if err != nil {
	//			fmt.Errorf("%v", err)
	//		}
	//	}
	//}
	occupied1181 := 0
	occupied1199 := 0
	applied1181 := 0
	applied1199 := 0
	for _, pod := range pods.Items {
		if pod.Annotations["ips"] != "" {
			switch pod.Annotations["location"] {
			case "1181":
				occupied1181++
			case "1199":
				occupied1199++
			}
		}
	}

	for ip := range ipToUsers {
		if strings.Contains(ip, "10.30") {
			applied1181++
		} else {
			applied1199++
		}
		idStr := ipToUsers[ip]
		findFlag := false
		for _, pod := range pods.Items {
			if strings.Contains(pod.Annotations["ips"], ip) {
				findFlag = true
				break
			}
		}

		id, _ := strconv.Atoi(idStr)

		if !findFlag {
			req := IpRelease{
				IP:     ip,
				UserId: id,
			}
			if ipToGroups[ip] != "" {
				req.Group = ipToGroups[ip]
			}
			fmt.Println(req)
			bytes, err := json.Marshal(req)
			if err != nil {
				fmt.Errorf("json errorf %v", err)
				continue
			}
			_, err = sendReleaseIpReq(bytes, "http://10.30.100.10:8090/api/net/ip/release")
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	fmt.Printf("1181: %v/%v, 1199:%v/%v", occupied1181, applied1181, occupied1199, applied1199)
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