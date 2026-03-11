package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

type BitacoraData struct {
	Name          string
	Account       string
	SocialNetwork string
	Channel       string
	UserName      string
	CreatedAt     string
	Category      string
	Description   string
	Area          string
	Colonia       string
	Priority      string
	Status        string
	Folio         string
	Observations  string
}

func main() {
	b, err := os.ReadFile("credentials.json")
	ctx := context.Background()

	jwtConfig, err := google.JWTConfigFromJSON(
		b,
		"https://www.googleapis.com/auth/spreadsheets.readonly",
	)
	if err != nil {
		log.Fatalf("Unable to parse service account file: %v", err)
	}

	client := jwtConfig.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetId := "1JHzuHqSx8eAmq77rcaPgfJLlTwVt9juzr1owtqvHlj0"
	readRange := "08 -14 de marzo"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		// sqlText := "INSERT INTO bitacora (created_at, area_id, category_id, channel_id, description, folio, link, observations, priority_id, status_id, username, colonia, social_network_id, available, user_id, account_id) VALUES "
		for _, row := range resp.Values {
			formatDate(row[6].(string))
			// values := fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)", row[0])
		}
	}
}

func formatDate(date string) {
	exampletime := time.Now()
	timeDate, err := time.Parse("2/1/2006", date)
	loc := exampletime.Location()
	mxTime := timeDate.In(loc)
	if err != nil {
		fmt.Println("Error")
	}
	fmt.Println(mxTime)
}

