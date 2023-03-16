package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.MaxMultipartMemory = 8 << 20

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/", func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Failed to upload image",
			})
			return
		}

		f, errOpen := file.Open()
		if errOpen != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": "Failed to upload image",
			})
			return
		}

		sess := ConnectAws()
		uploader := s3manager.NewUploader(sess)

		bucket := aws.String(os.Getenv("AWS_BUCKET"))

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: bucket,
			ACL:    aws.String("public-read"),
			Key:    aws.String(file.Filename),
			Body:   f,
		})
		if err != nil {
			c.HTML(http.StatusBadRequest, "index.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"image": "https://" + *bucket + ".s3." + os.Getenv("AWS_REGION") + "." + "amazonaws.com/" + file.Filename,
		})

		return
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func ConnectAws() *session.Session {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	myRegion := os.Getenv("AWS_REGION")
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(myRegion),
			Credentials: credentials.NewStaticCredentials(
				accessKeyID,
				secretAccessKey,
				"", // a token will be created when the session it's used.
			),
		})
	if err != nil {
		panic(err)
	}
	return sess
}
