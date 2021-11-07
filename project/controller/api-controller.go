package controller

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"project/geocode"
	"project/util"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CovidData struct {
	ID        string    `json:"_id" bson:"_id"`
	Date      time.Time `json:"date" bson:"date"`
	State     string    `json:"state" bson:"state"`
	Confirmed int64     `json:"confirmed" bson:"confirmed"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	Deceased  int64     `json:"deceased" bson:"deceased"`
	Recovered int64     `json:"recovered" bson:"recovered"`
	Tested    int64     `json:"tested" bson:"tested"`
}

// SyncDataWithSource
// @Summary Sync New Covid data
// @Description Downloads latest covid-19 data and updates the database
// @Tags root
// @Accept application/json
// @Produce plain
// @Success 200 {string} string	"Done syncing"
// @Failure 400 {string} string "Some error occurred"
// @Failure 404 {string} string "Data not found"
// @Router / [post]
func SyncDataWithSource(ctx echo.Context) error {

	client := http.Client{}

	csvResponse, requestError := client.Get("https://data.covid19india.org/csv/latest/states.csv")

	if requestError != nil {
		fmt.Println("unable to fetch csv" + requestError.Error())
		return requestError
	}

	fmt.Println("fetched csv")

	defer csvResponse.Body.Close()

	csvLines, readCsvError := csv.NewReader(csvResponse.Body).ReadAll()

	if readCsvError != nil {
		fmt.Println("unable to read csv lines" + readCsvError.Error())
		return readCsvError
	}

	var documents []mongo.WriteModel

	for _, line := range csvLines {
		date, parseDateError := util.ParseDateString(line[0])

		if parseDateError == nil {

			operation := mongo.NewUpdateOneModel()
			operation.SetFilter(bson.M{
				"$and": bson.A{
					bson.M{"state": line[1]},
					bson.M{"date": date},
				},
			})
			operation.SetUpdate(bson.M{
				"$set": bson.M{
					"confirmed":  util.ParseInteger64String(line[2]),
					"recovered":  util.ParseInteger64String(line[3]),
					"deceased":   util.ParseInteger64String(line[4]),
					"tested":     util.ParseInteger64String(line[6]),
					"created_at": time.Now(),
				},
			})
			operation.SetUpsert(true)

			documents = append(documents, operation)
		}

	}

	mongoConnectContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, connectDbError := mongo.Connect(mongoConnectContext, options.Client().ApplyURI("mongodb+srv://dbAdmin:ZMnacZrw9qUuIjuP@cluster0.foeab.mongodb.net/covid19india"))
	if connectDbError != nil {
		fmt.Println("unable to connect to database")
		return connectDbError
	}

	database := mongoClient.Database("covid19india")
	collection := database.Collection("covid_data")

	insertManyContext, cancel := context.WithDeadline(context.Background(), time.Now().Add(20*time.Second))
	defer cancel()
	_, bulkWriteError := collection.BulkWrite(insertManyContext, documents, nil)

	if bulkWriteError != nil {
		fmt.Println("unable to bulk write" + bulkWriteError.Error())
		return bulkWriteError
	}

	disconnectMongoClientError := mongoClient.Disconnect(mongoConnectContext)
	if disconnectMongoClientError != nil {
		fmt.Println("unable to disconnect" + disconnectMongoClientError.Error())
		return disconnectMongoClientError
	}

	return ctx.String(http.StatusOK, "Done syncing\n")
}

// GetDataForLocation
// @Summary Get my state data
// @Description Returns the latest covid-19 data for user's state and for India
// @Tags root
// @Accept */*
// @Produce json
// @Param lat query number true "User's latitude"
// @Param long query number true "User's Longitude"
// @Success 200 {object} []interface{}
// @Failure 400 {string} string "Some error occurred"
// @Failure 404 {string} string "Data not found"
// @Router / [get]
func GetDataForLocation(ctx echo.Context) error {
	latitude := util.ParseFloat64String(ctx.QueryParam("lat"))
	longitude := util.ParseFloat64String(ctx.QueryParam("long"))

	if latitude == -1 || longitude == -1 {
		return ctx.String(http.StatusBadRequest, "Bad Longitude/Latitude\n")
	}

	state := geocode.RevGeocodeState(latitude, longitude)

	mongoConnectContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, connectDbError := mongo.Connect(mongoConnectContext, options.Client().ApplyURI("mongodb+srv://dbAdmin:ZMnacZrw9qUuIjuP@cluster0.foeab.mongodb.net/covid19india"))
	if connectDbError != nil {
		fmt.Println("unable to connect to database")
		return connectDbError
	}

	database := mongoClient.Database("covid19india")
	collection := database.Collection("covid_data")

	findOneContext, cancel := context.WithDeadline(context.Background(), time.Now().Add(20*time.Second))
	defer cancel()
	findStateResult := collection.FindOne(findOneContext, bson.M{"state": state}, options.FindOne().SetSort(bson.M{"date": -1}))
	if findStateResult.Err() != nil {
		return findStateResult.Err()
	}
	findIndiaResult := collection.FindOne(findOneContext, bson.M{"state": "India"}, options.FindOne().SetSort(bson.M{"date": -1}))
	if findIndiaResult.Err() != nil {
		return findIndiaResult.Err()
	}

	disconnectMongoClientError := mongoClient.Disconnect(mongoConnectContext)
	if disconnectMongoClientError != nil {
		fmt.Println("unable to disconnect" + disconnectMongoClientError.Error())
		return disconnectMongoClientError
	}

	var stateCovidData CovidData
	decodeStateResult := findStateResult.Decode(&stateCovidData)
	if decodeStateResult != nil {
		return decodeStateResult
	}

	var centralCovidData CovidData
	decodeCentralResult := findIndiaResult.Decode(&centralCovidData)
	if decodeCentralResult != nil {
		return decodeCentralResult
	}

	return ctx.JSON(http.StatusOK, []interface{}{stateCovidData, centralCovidData})

}
