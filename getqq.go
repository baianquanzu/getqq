package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// 读取HTML文件
func readHTMLFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// 提取群名称
func extractGroupName(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var groupName string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, a := range n.Attr {
				if a.Key == "id" && a.Val == "groupTit" {
					if n.FirstChild != nil {
						groupName = n.FirstChild.Data
					}
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	if groupName == "" {
		return "", fmt.Errorf("未能提取到群名称")
	}
	return groupName, nil
}

// 提取QQ号并去重
func extractUniqueQQNumbers(htmlContent string) []string {
	qqSet := make(map[string]struct{})
	re := regexp.MustCompile(`\b[1-9][0-9]{4,11}\b`)
	matches := re.FindAllString(htmlContent, -1)

	for _, match := range matches {
		qqSet[match] = struct{}{}
	}

	var qqNumbers []string
	for qq := range qqSet {
		qqNumbers = append(qqNumbers, qq)
	}

	return qqNumbers
}

// 保存结果到文件
func saveToFile(qqNumbers []string, groupName string) error {
	// 处理群名称，确保文件名合法
	fileName := strings.ReplaceAll(groupName, " ", "_") + ".txt"
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("群名称: %s\n", groupName))
	if err != nil {
		return err
	}

	for _, qqNumber := range qqNumbers {
		_, err = file.WriteString(fmt.Sprintf("%s\n", qqNumber))
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// 指定要读取的文件夹路径
	folderPath := "./html_files"

	// 检查文件夹是否存在
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		log.Fatalf("文件夹 %s 不存在，请创建该文件夹并放入HTML文件", folderPath)
	}

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			fmt.Printf("处理文件: %s\n", path)
			htmlContent, err := readHTMLFile(path)
			if err != nil {
				log.Fatalf("读取HTML文件失败: %v", err)
			}

			groupName, err := extractGroupName(htmlContent)
			if err != nil {
				log.Fatalf("提取群名称失败: %v", err)
			}

			qqNumbers := extractUniqueQQNumbers(htmlContent)

			fmt.Println("提取的QQ号:")
			for _, qqNumber := range qqNumbers {
				fmt.Println(qqNumber)
			}

			err = saveToFile(qqNumbers, groupName)
			if err != nil {
				log.Fatalf("保存到文件失败: %v", err)
			}

			fmt.Printf("结果已保存到 %s.txt 文件中\n", strings.ReplaceAll(groupName, " ", "_"))
		}
		return nil
	})

	if err != nil {
		log.Fatalf("遍历文件夹失败: %v", err)
	}
}
