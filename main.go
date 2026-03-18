package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	readRange := "06-11 feb"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		// sqlText := "INSERT INTO bitacora (created_at, area_id, category_id, channel_id, description, folio, link, observations, priority_id, status_id, username, colonia, social_network_id, available, user_id, account_id) VALUES "
		var values []string
		for _, row := range resp.Values {
			var creatredByName string
			if len(row) == 0 {
				continue
			}

			// skip header row if needed
			first := strings.TrimSpace(fmt.Sprint(row[0]))
			if first == "Cuenta" || first == "ACCOUNT" || first == "Responsable" || first == "Nombre" {
				continue
			}

			getCell := func(i int) string {
				if i >= len(row) {
					return ""
				}
				return strings.TrimSpace(fmt.Sprint(row[i]))
			}

			// Cells without UserID
			accountText := getCell(0)
			socialNetworkText := getCell(1)
			channelText := getCell(2)
			username := getCell(3)
			createdAt := getCell(4)
			categoryText := getCell(5)
			description := getCell(6)
			areaText := getCell(7)
			colonia := getCell(8)
			priorityText := getCell(9)
			statusText := getCell(10)
			folio := getCell(11)
			observations := getCell(12)
			/*
				userName := getCell(0)
				accountText := getCell(1)
				socialNetworkText := getCell(2)
				channelText := getCell(3)
				username := getCell(4)
				link := getCell(5)
				createdAt := getCell(6)
				categoryText := getCell(7)
				description := getCell(8)
				areaText := getCell(9)
				colonia := getCell(10)
				priorityText := getCell(11)
				statusText := getCell(12)
				folio := getCell(13)
				observations := getCell(14)
			*/
			// normalize only if your maps use normalized keys
			accountKey := formatLowerDash(accountText)
			socialNetworkKey := formatLowerDash(socialNetworkText)
			channelKey := channelText
			categoryKey := formatLowerDash(categoryText)
			areaKey := formatLowerDash(areaText)
			priorityKey := formatLowerDash(priorityText)
			statusKey := formatLowerDash(statusText)
			// userId := MapOptionalUserID(userIdMap, userName)
			accountID := MapOptional(accountMap, accountKey)
			socialNetworkID := MapOptional(socialNetworkMap, socialNetworkKey)
			channelID := MapOptional(channelMap, channelKey)
			categoryID := MapOptional(categoryMap, categoryKey)
			areaID := MapOptional(areaMap, areaKey)
			priorityID := MapOptional(priorityMap, priorityKey)
			statusID := MapOptional(statusMap, statusKey)
			// fields not coming from sheet yet
			link := "NULL"
			available := true
			userID := "NULL"
			// if userId == "" {
			// 	creatredByName = userName
			// }
			valueRow := fmt.Sprintf(
				"(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
				formatDate(createdAt),            // created_at
				sqlIntPtr(areaID),                // area_id
				sqlIntPtr(categoryID),            // category_id
				sqlIntPtr(channelID),             // channel_id
				sqlString(description),           // description
				sqlString(folio),                 // folio
				link,                             // link
				sqlString(observations),          // observations
				sqlIntPtr(priorityID),            // priority_id
				sqlIntPtr(statusID),              // status_id
				sqlString(username),              // username
				sqlString(colonia),               // colonia
				sqlIntPtr(socialNetworkID),       // social_network_id
				sqlBool(available),               // available
				userID,                           // user_id
				sqlIntPtr(accountID),             // account_id
				sqlStrigsPointer(creatredByName), // created_by_name
				formatDate(createdAt),            // updated_at is the same as created_at (initially)
			)
			values = append(values, valueRow)
		}
		// Without userId & link
		// sqlText := "INSERT INTO bitacora (created_at, area_id, category_id, channel_id, description, folio, link, observations, priority_id, status_id, username, colonia, social_network_id, available, user_id, account_id) VALUES " + strings.Join(values, ",\n")
		sqlText := "INSERT INTO bitacora (created_at, area_id, category_id, channel_id, description, folio, link, observations, priority_id, status_id, username, colonia, social_network_id, available, user_id, account_id, created_by_name, updated_at) VALUES " + strings.Join(values, ",\n")
		formatedText := []byte(fmt.Sprintf("%s;", sqlText))
		path := filepath.Join(os.TempDir(), "test.sql")
		err := os.WriteFile(path, formatedText, 0644)
		check(err)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func sqlStrigsPointer(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "NULL"
	}
	s = strings.ReplaceAll(s, "'", "''")
	return fmt.Sprintf("'%s'", s)
}

func sqlString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "NULL"
	}
	s = strings.ReplaceAll(s, "'", "''")
	return fmt.Sprintf("'%s'", s)
}

func sqlIntPtr(n *int) string {
	if n == nil {
		return "NULL"
	}
	return fmt.Sprintf("%d", *n)
}

func sqlBool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

func formatLowerDash(text string) string {
	lowerCaseText := strings.ToLower(text)
	splitText := strings.Split(lowerCaseText, " ")
	joinText := strings.Join(splitText, "_")
	return joinText
}

func MapOptionalUserID(dict map[string]string, key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

	id, ok := dict[key]
	if !ok {
		return ""
	}
	return id
}

func MapOptional(dict map[string]int, key string) *int {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}

	id, ok := dict[key]
	if !ok {
		return nil
	}

	return &id
}

var userIdMap = map[string]string{
	"Sayra":  "03f6bcc8-9ad3-490f-95d7-3fbbc30ef720",
	"Naeli":  "a255cc89-f9de-46df-a67c-5c9fab8ba29e",
	"Declan": "05b3eb0c-302c-4fe6-8eff-3c9ea4d0a27f",
	"Zamah":  "e7ed810b-f2d6-4a0d-96d5-6db55a950668",
}

var accountMap = map[string]int{
	"zap": 0,
	"jjf": 1,
}

var socialNetworkMap = map[string]int{
	"facebook":  1,
	"x":         2,
	"instagram": 3,
	"tiktok":    4,
}

var areaMap = map[string]int{
	"infraestructura_de_comercio":  0,
	"servicios_municipales":        1,
	"gestión_integral":             2,
	"secretaría_del_ayuntamiento":  3,
	"desarrollo_económico":         4,
	"construcción_comunidad":       5,
	"dif":                          6,
	"tesorería":                    7,
	"cfe":                          8,
	"siapa":                        9,
	"siop":                         10,
	"otras_coordinaciones":         11,
	"otras_dependencias_estatales": 12,
	"presidencia":                  13,
	"guadalajara":                  14,
	"inspección_y_vigilancia":      15,
	"pcyb":                         16,
	"cercanía_ciudadana":           17,
	"salud_zapopan":                18,
	"comisaría":                    19,
	"comude":                       20,
	"caec_(boletos_charros)":       21,
	"sindicatura":                  22,
	"administración_e_innovación_gubernamental": 23,
	"amim": 24,
	"cursos_en_el_parque_de_las_niñas_y_niños": 25,
	"romería":               26,
	"contraloría_ciudadana": 27,
	"toc_toc":               28,
	"otros":                 29,
	"equipo_campaña":        30,
	"fiesta_de_abril":       31,
	"desabasto_de_agua_en_lomas_de_centinela": 32,
	"infraestrucura_en_comercio":              33,
}

var categoryMap = map[string]int{
	"solicitud_de_información":           0,
	"canalización_a_dependencia":         1,
	"solicitudes_nuevas":                 2,
	"reportes_de_servicios":              3,
	"reportes_de_obras":                  4,
	"reportes_externos":                  5,
	"solicitudes_especiales":             6,
	"reporte_de_inspección_y_vigilancia": 7,
	"reportes_y_denuncias":               8,
	"solicitud_de_empleo":                9,
	"coyuntura":                          10,
	"participación_en_curso":             11,
	"solicitud_de_obra":                  12,
	"otros":                              13,
}

var channelMap = map[string]int{
	"Comentario": 0,
	"Inbox":      1,
}

var priorityMap = map[string]int{
	"baja":  0,
	"media": 1,
	"alta":  2,
}

var statusMap = map[string]int{
	"pendiente":  0,
	"en_proceso": 1,
	"resuelto":   2,
	"dirección":  3,
}

func formatDate(date string) string {
	loc, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		log.Println("load location error:", err)
		return ""
	}
	timeDate, err := time.ParseInLocation("2/1/2006", date, loc)
	if err != nil {
		fmt.Println("Error")
	}

	return fmt.Sprintf("'%s'", timeDate.Format("2006-01-02 15:04:05-07:00"))
}
