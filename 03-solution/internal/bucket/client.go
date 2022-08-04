package bucket

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"solution/internal/model"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/notification"
)

const BucketForData = "datalake"
const BucketForResult = "processed-data"
const LocalPathToResult = "processed_data/out.csv"

type FileObject struct {
	Key, Name, Extension string
}

type ClientCan interface {
	GetAllObjects() ([]FileObject, error)
	AggregateUsers([]FileObject) (model.Users, error)
	StoreUsers(model.Users, string) error
	FilterUsers(model.Users, map[string]string) (model.Users, error)
}

type Client struct {
	client     *minio.Client
	bucketName string
	UpdateCh   chan struct{}
	ticker     *time.Ticker
}

func NewFileObject(fileName string) FileObject {
	return FileObject{
		Key:       fileName,
		Name:      strings.Replace(fileName, filepath.Ext(fileName), "", 1),
		Extension: filepath.Ext(fileName),
	}
}

func NewClient(serverAddr, accessKeyID, secretAccessKey string) *Client {
	minioClient, err := minio.New(serverAddr, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		log.Fatalln("Error initializing minio client -\t", err)
	}
	return &Client{client: minioClient, ticker: time.NewTicker(5 * time.Minute), UpdateCh: make(chan struct{})}
}

func (bc *Client) SetBucket(bucket string) error {
	ctx := context.TODO()
	ok, err := bc.client.BucketExists(ctx, bucket)
	if err != nil {
		err = fmt.Errorf("error getting access to bucket %s -\t%w\n", bucket, err)
		return err
	}
	if !ok {
		err = bc.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			err = fmt.Errorf("error creating new bucket %s -\t%w\n", bucket, err)
			return err
		}
	}

	bc.bucketName = bucket
	return nil
}

func (bc *Client) GetAllObjects() ([]FileObject, error) {
	objectCh := bc.client.ListObjects(context.TODO(), bc.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})
	result := make([]FileObject, 0, 10)
	for obj := range objectCh {
		if obj.Err != nil {
			log.Println("Error getting object from the bucket:\t", obj.Err)
			return result, obj.Err
		}
		result = append(result, NewFileObject(obj.Key))
	}

	return result, nil
}

func (bc *Client) getObjectData(fo FileObject) io.Reader {
	obj, err := bc.client.GetObject(context.TODO(), bc.bucketName, fo.Key, minio.GetObjectOptions{})
	if err != nil {
		log.Println("Error reading file data of\t", fo.Key)
		return nil
	}
	return obj
}

func (bc *Client) getUserData(fo FileObject) *model.User {
	if strings.ToLower(fo.Extension) != ".csv" {
		return nil
	}

	objData := bc.getObjectData(fo)
	if objData == nil {
		return nil
	}

	dataCSV, err := csv.NewReader(objData).ReadAll()
	if err != nil {
		log.Println("Error reading csv data\t", fo.Key, "\t", err)
		return nil
	}

	if len(dataCSV) < 2 {
		log.Println("Wrong csv format\t", fo.Key)
		return nil
	}

	userDataStr := dataCSV[1]
	if len(userDataStr) < 3 {
		log.Println("Wrong user string format in\t", fo.Key)
		return nil
	}

	return &model.User{
		ID:        fo.Name,
		Firstname: strings.Trim(userDataStr[0], " \t"),
		Lastname:  strings.Trim(userDataStr[1], " \t"),
		Birthday:  timeFromMillis(userDataStr[2]),
	}
}

func timeFromMillis(strTime string) time.Time {
	millisec, err := strconv.ParseInt(strings.Trim(strTime, " \t"), 10, 64)
	if err != nil {
		log.Println("time conversion error\t", strTime, "\t", err)
		return time.Time{}
	}

	return time.UnixMilli(millisec)
}

func (bc *Client) GetUserFromObject(fo FileObject) *model.User {
	if strings.ToLower(fo.Extension) == ".png" {
		return &model.User{
			ID:          fo.Name,
			PicturePath: fo.Key,
		}
	}
	user := bc.getUserData(fo)
	if user == nil {
		log.Println("Can't get user from object\t", fo.Key)
		return nil
	}
	return user
}

func combineUsers(userWas, userNew *model.User) *model.User {
	switch {
	case userWas == nil && userNew == nil:
		return &model.User{}
	case userWas == nil:
		return userNew
	case userNew == nil:
		return userWas
	}

	result := &model.User{
		ID:          getNonEmptyS(userNew.ID, userWas.ID),
		Firstname:   getNonEmptyS(userNew.Firstname, userWas.Firstname),
		Lastname:    getNonEmptyS(userNew.Lastname, userWas.Lastname),
		Birthday:    getNonEmptyT(userNew.Birthday, userWas.Birthday),
		PicturePath: getNonEmptyS(userNew.PicturePath, userWas.PicturePath),
	}
	return result
}

func getNonEmptyS(str ...string) string {
	for _, v := range str {
		if len(v) > 0 {
			return v
		}
	}
	return ""
}

func getNonEmptyT(tim ...time.Time) time.Time {
	for _, v := range tim {
		if !v.IsZero() {
			return v
		}
	}
	return time.Time{}
}

func (bc *Client) AggregateUsers(objects []FileObject) (model.Users, error) {
	result := make(model.Users, len(objects))
	for _, obj := range objects {
		objUser := bc.GetUserFromObject(obj)
		result[obj.Name] = combineUsers(objUser, result[obj.Name])
	}

	return result, nil
}

func (bc *Client) StoreUsers(users model.Users, pathToStore string) error {
	err := bc.SetBucket(BucketForResult)
	if err != nil {
		log.Println("Unable to set bucket for result:\t", BucketForResult)
		return err
	}
	defer bc.SetBucket(BucketForData)

	csvFile, err := os.Create(pathToStore)
	if err != nil {
		log.Println("Error creating result file -\t", err)
		return err
	}

	curResult := []string{"user_id", "first_name", "last_name", "births", "img_path"}
	csvWriter := csv.NewWriter(csvFile)
	err = csvWriter.Write(curResult)
	if err != nil {
		return err
	}
	for _, user := range users {
		curResult[0] = user.ID
		curResult[1] = user.Firstname
		curResult[2] = user.Lastname
		curResult[3] = user.Birthday.String()
		curResult[4] = user.PicturePath
		err = csvWriter.Write(curResult)
		if err != nil {
			log.Println("error writing to csv-result\t", err)
		}
	}
	csvWriter.Flush()

	_, err = bc.client.FPutObject(context.TODO(), bc.bucketName, filepath.Base(pathToStore), pathToStore, minio.PutObjectOptions{ContentType: "application/csv"})

	return err
}

func (bc *Client) FilterUsers(users model.Users, filters map[string]string) (model.Users, error) {
	if filters == nil {
		return users, nil
	}

	howToFilter := model.NewUserFilter()
	err := howToFilter.AdjustFilters(filters)
	if err != nil {
		return users, err
	}

	result := make(model.Users)
	for id, user := range users {
		if howToFilter.FitsToAll(user) {
			result[id] = user
		}
	}

	return result, nil
}

func (bc *Client) SetBucketListener(ctx context.Context, bucketName string, forUsers model.Users) error {
	queueArn := notification.NewArn("minio", "sqs", "", "1", "nsq")
	queueConfig := notification.NewConfig(queueArn)
	queueConfig.AddEvents(notification.ObjectCreatedAll, notification.ObjectRemovedAll)
	config := notification.Configuration{}
	config.AddQueue(queueConfig)
	err := bc.client.SetBucketNotification(ctx, bucketName, config)
	if err != nil {
		return fmt.Errorf("unable to set bucket notification:\t%w", err)
	}
	go bc.catchBucketNotifications(ctx, bucketName, forUsers)
	return nil
}

func (bc *Client) catchBucketNotifications(ctx context.Context, bucketName string, forUsers model.Users) {
	notificationChannel := bc.client.ListenBucketNotification(ctx, bucketName, "", "",
		[]string{
			"s3:ObjectCreated:*",
			"s3:ObjectRemoved:*"})
	changedUsersData := make(model.Users)

	go func() {
		for {
			select {
			case <-bc.UpdateCh:
				log.Println("Force bucket data update:\t", bucketName)
				err := bc.saveChanges(forUsers, changedUsersData)
				if err != nil {
					log.Println(err)
				}
			case <-bc.ticker.C:
				err := bc.saveChanges(forUsers, changedUsersData)
				if err != nil {
					log.Println(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for notificationInfo := range notificationChannel {
		if notificationInfo.Err != nil {
			continue
		}

		for _, record := range notificationInfo.Records {
			objectName := record.S3.Object.Key
			fo := NewFileObject(objectName)
			_, ok := changedUsersData[fo.Name]
			if !ok {
				changedUsersData[fo.Name] = combineUsers(changedUsersData[fo.Name], forUsers[fo.Name])
			}
			if strings.Contains(record.EventName, "s3:ObjectRemoved") {
				if changedUsersData[fo.Name] == nil {
					continue
				}
				if strings.ToLower(fo.Extension) != ".csv" {
					changedUsersData[fo.Name].PicturePath = ""
				} else {
					changedUsersData[fo.Name] = nil
				}
				continue
			}
			changedUser := bc.GetUserFromObject(fo)
			changedUsersData[fo.Name] = combineUsers(changedUsersData[fo.Name], changedUser)
		}
	}
}

func (bc *Client) saveChanges(dataMap, changedUsersData model.Users) error {
	for key := range changedUsersData {
		if changedUsersData[key] == nil {
			delete(dataMap, key)
			continue
		}
		dataMap[key] = changedUsersData[key]
	}
	err := bc.StoreUsers(dataMap, LocalPathToResult)
	if err != nil {
		return fmt.Errorf("store users after changes made in the bucket\t%w", err)
	}
	changedUsersData = make(model.Users)
	return nil
}
