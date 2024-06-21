package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
)

type Book struct {
	Author            string `json:"author"`
	Cover             string `json:"cover"`
	Title             string `json:"operating_title"`
	AuthorInfo        string `json:"author_info"`
	BookIntro         string `json:"book_intro"`
	PublishTime       string `json:"publish_time"`
	Uptime            string `json:"uptime"`
	OtherShareSummary string `json:"other_share_summary"`
	Enid              string `json:"enid"` // dedao 每本书的唯一标识
}

type Response struct {
	List []Book `json:"list"`
}

func initLogger() *log.Logger {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		os.Exit(1)
	}
	logger := log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	return logger
}

var logger = initLogger()
var initialCreatedTime time.Time

func init() {
	// 初始化时设置创建时间
	initialCreatedTime = time.Now()
}

// fetchBooks 从得到获取电子书列表
func fetchBooks() ([]Book, error) {
	url := "https://m.igetget.com/native/api/ebook/getBookList"
	payload := `{"count": 50, "max_id": 0, "sort": "time", "since_id": 0}`

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		logger.Println("Error creating request:", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		C Response `json:"c"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Println("Error decoding response:", err)
		return nil, err
	}

	return result.C.List, nil
}

// generateAtom 生成 Atom 订阅源
func generateAtom(books []Book) (string, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "得到最新电子书 Atom 订阅源",
		Link:        &feeds.Link{Href: "https://m.igetget.com/native/ebook/#/ebook/newBookList"},
		Description: "得到最新电子书更新",
		Created:     initialCreatedTime,
		Updated:     now,
	}

	// book detail url: https://www.dedao.cn/ebook/detail?id=enid
	for _, book := range books {
		createdTime, err := time.Parse("2006-01-02 15:04:05", book.Uptime)
		if err != nil {
			logger.Println("Error parsing time:", err)
			continue
		}

		item := &feeds.Item{
			Title: book.Title,
			Link:  &feeds.Link{Href: fmt.Sprintf("https://www.dedao.cn/ebook/detail?id=%s", book.Enid)},
			Content: fmt.Sprintf(
				"<img src='%s'/><br/>主编推荐语: %s<br/><br/>内容简介: %s<br/><br/>作者介绍: %s<br/><br/>出版时间: %s",
				book.Cover,
				book.OtherShareSummary,
				book.BookIntro,
				book.AuthorInfo,
				book.PublishTime,
			),
			Author:      &feeds.Author{Name: book.Author},
			Created:     createdTime,
			Id:          book.Enid,
			Description: book.OtherShareSummary,
		}
		feed.Items = append(feed.Items, item)
	}

	atom, err := feed.ToAtom()
	if err != nil {
		logger.Println("Error generating Atom feed:", err)
		return "", err
	}

	return atom, nil
}

// saveAtomToFile 将 Atom 订阅源保存到文件
func saveAtomToFile(atom string) error {
	file, err := os.Create("dedao.atom")
	if err != nil {
		logger.Println("Error creating Atom file:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(atom)
	if err != nil {
		logger.Println("Error writing to Atom file:", err)
		return err
	}

	return nil
}

// updateAtomFile 更新 Atom 订阅源
func updateAtomFile() {
	defer func() {
		if r := recover(); r != nil {
			logger.Println("Panic Recovered in updateAtomFile:", r)
		}
	}()

	for {
		books, err := fetchBooks()
		if err != nil {
			logger.Println("Error fetching books:", err)
			continue
		}

		atom, err := generateAtom(books)
		if err != nil {
			logger.Println("Error generating Atom feed:", err)
			continue
		}

		if err := saveAtomToFile(atom); err != nil {
			logger.Println("Error saving Atom file:", err)
			continue
		}

		logger.Println("Atom feed successfully updated")
		time.Sleep(2 * time.Hour)
	}
}

func main() {
	go updateAtomFile()

	r := gin.Default()

	r.GET("/feeds/dedao.atom", func(c *gin.Context) {
		data, err := os.ReadFile("dedao.atom")
		if err != nil {
			logger.Println("Error reading Atom file:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Read Atom file failed"})
			return
		}

		c.Header("Content-Type", "application/atom+xml; charset=utf-8")
		c.Data(http.StatusOK, "application/atom+xml; charset=utf-8", data)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	// 优雅重启
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Println("Server exiting")
}
