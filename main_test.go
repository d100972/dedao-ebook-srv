package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFetchBooks(t *testing.T) {
	books, err := fetchBooks()
	if err != nil {
		t.Fatalf("fetchBooks failed: %v", err)
	}

	if len(books) != 50 {
		t.Fatalf("expected 50 books, got %d", len(books))
	}
}

func TestGenerateAtom(t *testing.T) {
	books := []Book{
		{
			Author:            "Author1",
			Cover:             "http://example.com/cover1.jpg",
			Title:             "Title1",
			AuthorInfo:        "Author Info 1",
			BookIntro:         "Book Intro 1",
			PublishTime:       "2023-01-01",
			Uptime:            "2023-01-01 00:00:00",
			OtherShareSummary: "Summary 1",
			Enid:              "enid1",
		},
	}

	atom, err := generateAtom(books)
	if err != nil {
		t.Fatalf("generateAtom failed: %v", err)
	}

	assert.Contains(t, atom, "<title>Title1</title>")
	assert.Contains(t, atom, "http://example.com/cover1.jpg")
}

func TestSaveAtomToFile(t *testing.T) {
	atom := "<feed><title>Test Feed</title></feed>"
	err := saveAtomToFile(atom)
	if err != nil {
		t.Fatalf("saveAtomToFile failed: %v", err)
	}

	data, err := os.ReadFile("dedao.atom")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	assert.Equal(t, atom, string(data))
}

func TestUpdateAtomFile(t *testing.T) {
	go updateAtomFile()

	time.Sleep(5 * time.Second)

	data, err := os.ReadFile("dedao.atom")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	assert.Contains(t, string(data), "<feed xmlns=\"http://www.w3.org/2005/Atom\">")
}

func TestMainRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.GET("/feeds/dedao.atom", func(c *gin.Context) {
		data, err := os.ReadFile("dedao.atom")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Read Atom file failed"})
			return
		}

		c.Header("Content-Type", "application/atom+xml; charset=utf-8")
		c.Data(http.StatusOK, "application/atom+xml; charset=utf-8", data)
	})

	req, _ := http.NewRequest("GET", "/feeds/dedao.atom", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "<feed xmlns=\"http://www.w3.org/2005/Atom\">")
}
