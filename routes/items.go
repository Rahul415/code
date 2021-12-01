package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"fmt"

	"voiceassistant/models"

	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v4"
)

func ItemsIndex(c *gin.Context) {
	db, _ := c.Get("db")
	conn := db.(pgx.Conn)
	items, err := models.GetAllItems(&conn)
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func ItemsCreate(c *gin.Context) {
	userID := c.GetString("user_id")
	db, _ := c.Get("db")
	conn := db.(pgx.Conn)

	item := models.Item{}
	c.ShouldBindJSON(&item)
	err := item.Create(&conn, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func ItemsForSaleByCurrentUser(c *gin.Context) {
	userID := c.GetString("user_id")
	db, _ := c.Get("db")
	conn := db.(pgx.Conn)

	items, err := models.GetItemsBeingSoldByUser(userID, &conn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func ItemsUpdate(c *gin.Context) {
	userID := c.GetString("user_id")
	db, _ := c.Get("db")
	conn := db.(pgx.Conn)

	itemSent := models.Item{}
	err := c.ShouldBindJSON(&itemSent)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form sent"})
		return
	}

	itemBeingUpdated, err := models.FindItemById(itemSent.ID, &conn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if itemBeingUpdated.ProductId.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this item"})
		return
	}

	itemSent.ProductId = itemBeingUpdated.ProductId
	err = itemSent.Update(&conn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": itemSent})
}

type FlaskOut struct {
	Keyword []string `json:"keyword"`
}

func SpeechToText(c *gin.Context) {

	var b bytes.Buffer
	var audio bytes.Buffer
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		return
	}
	_, err = io.Copy(&audio, file)
	if err != nil {
		return
	}

	w := multipart.NewWriter(&b)

	flask_audio, err := w.CreateFormFile("file", "audio.wav")
	if err != nil {
		return
	}
	_, err = io.Copy(flask_audio, &audio)
	if err != nil {
		return
	}

	w.Close()

	req, err := http.NewRequest("POST", "http://34.92.72.160:8080/gettext", &b)

	if err != nil {
		return
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {

		return
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errMsg := "unable to read response data"
		log.Println("Msg: " + errMsg + " Error: " + err.Error())
		return
	}

	myJson := string(responseData)
	jsonConfig := []byte(myJson)
	var Out FlaskOut
	err = json.Unmarshal(jsonConfig, &Out)
	if err != nil {
		errMsg := "unable to unmarshal http response"
		log.Println("Msg: " + errMsg + " Error: " + err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "error",
		})
	} else {
		c.JSON(http.StatusOK, map[string]interface{}{
			"message":  "Data is inserted succesfully",
			"products": Out.Keyword,
		})
	}
}

// func with out.keyword[0] return 2 vars string and err
//speech from flask and text  inpput to another api by var
//

func Emotion(c *gin.Context) {
	text := c.Request.PostFormValue("itemloc")
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("text", text)
	w.Close()

	req, err := http.NewRequest("POST", "http://34.92.72.160:8080/getspeech", &b)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatalf("ioutil.ReadAll -> %v", err)
	}

	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Println(string(data))
	} else {
		log.Println("error")
	}
}
