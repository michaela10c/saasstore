package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"reflect"

	"appstore/backend"
	"appstore/constants"
	"appstore/gateway/stripe"
	"appstore/model"

	"github.com/olivere/elastic/v7"
)


func SearchApps(title string, description string) ([]model.App, error) {
   if title == "" {
       return SearchAppsByField("description", description)
   }
   if description == "" {
       return SearchAppsByField("title", title)
   }

   // 1. construct search query
   query1 := elastic.NewMatchQuery("title", title)
   query1.Operator("AND")
   query2 := elastic.NewMatchQuery("description", description)
   query2.Operator("AND")
   query := elastic.NewBoolQuery().Must(query1, query2)

   // 2. call backend
   searchResult, err := backend.ESBackend.ReadFromES(query, constants.APP_INDEX)
   
   if err != nil {
       return nil, err
   }

   // 3. process result
   return getAppFromSearchResult(searchResult), nil
}

func SearchAppsByField(field string, value string) ([]model.App, error) {
   // 1. construct search query
   query := elastic.NewMatchQuery(field, value)
   query.Operator("AND")
   if field == "" {
       query.ZeroTermsQuery("all")
   }

   // 2. call backend
   searchResult, err := backend.ESBackend.ReadFromES(query, constants.APP_INDEX)
   if err != nil {
       return nil, err
   }

   // 3. process result
   return getAppFromSearchResult(searchResult), nil
}

func getAppFromSearchResult(searchResult *elastic.SearchResult) []model.App {
   var ptype model.App
   var apps []model.App
   for _, item := range searchResult.Each(reflect.TypeOf(ptype)) {
       p := item.(model.App)
       apps = append(apps, p)
   }
   return apps
}

 func SaveApp(app *model.App, file multipart.File) error {
    // 1. create price & product in stripe 
    productID, priceID, err := stripe.CreateProductWithPrice(app.Title, app.Description, int64(app.Price*100))
    if err != nil {
        fmt.Printf("Failed to create Product and Price using Stripe SDK %v\n", err)
        return err
    }

    // update productID + price ID to the app 
    app.ProductID = productID
    app.PriceID = priceID
    fmt.Printf("Product %s with price %s is successfully created", productID, priceID)

    // save uploaded App data to GCS
    medialink, err := backend.GCSBackend.SaveToGCS(file, app.Id)
    if err != nil {
        return err
    }
    app.Url = medialink
    
    // save uploaded App data to ES
    err = backend.ESBackend.SaveToES(app, constants.APP_INDEX, app.Id)
    if err != nil {
        fmt.Printf("Failed to save app to elastic search with app index %v\n", err)
        return err
    }
    fmt.Println("App is saved successfully to ES app index.")
 
    return nil
 }
 
 

 func SearchAppByID(appID string) (*model.App, error) {
    // 1. construct search query
    query := elastic.NewTermQuery("id", appID)

    // 2. call backend
    searchResult, err := backend.ESBackend.ReadFromES(query, constants.APP_INDEX)
    if err != nil {
        return nil, err
    }

    // 3. process result
    results := getAppFromSearchResult(searchResult)
    if len(results) == 1 {
        return &results[0], nil
    }

    // error handling
    return nil, nil
 }
 

 func CheckoutApp(domain string, appID string) (string, error) {
    //appID --> priceID
    app, err := SearchAppByID(appID)
    if err != nil {
        return "", err
    }
    if app == nil {
        return "", errors.New("unable to find app in elasticsearch")
    }

    // Create a checkout session by passing in PriceID
    return stripe.CreateCheckoutSession(domain, app.PriceID)
 }
 
 func DeleteApp(id string, user string) error {
    query := elastic.NewBoolQuery()
    query.Must(elastic.NewTermQuery("id", id))
    query.Must(elastic.NewTermQuery("user", user))
    
    return backend.ESBackend.DeleteFromES(query, constants.APP_INDEX)
 }
