package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type DDLData struct {
	Candidates []struct {
		Index   int `json:"Index"`
		Content struct {
			Parts []string `json:"Parts"`
			Role  string   `json:"Role"`
		} `json:"Content"`
		FinishReason  int `json:"FinishReason"`
		SafetyRatings []struct {
			Category    int  `json:"Category"`
			Probability int  `json:"Probability"`
			Blocked     bool `json:"Blocked"`
		} `json:"SafetyRatings"`
		CitationMetadata interface{} `json:"CitationMetadata"`
		TokenCount       int         `json:"TokenCount"`
	} `json:"Candidates"`
	PromptFeedback interface{} `json:"PromptFeedback"`
	UsageMetadata  struct {
		PromptTokenCount        int `json:"PromptTokenCount"`
		CachedContentTokenCount int `json:"CachedContentTokenCount"`
		CandidatesTokenCount    int `json:"CandidatesTokenCount"`
		TotalTokenCount         int `json:"TotalTokenCount"`
	} `json:"UsageMetadata"`
}

type DummyDataResult struct {
	Inserts    string `json:"inserts"`
	CreateJSON string `json:"create_json"`
	UpdateJSON string `json:"update_json"`
}

func GenerateDummyData(class string, classMetadata [][]string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatalf("GEMINI_API_KEY not found in environment variables")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")

	var formattedMetadata []string
	for _, pair := range classMetadata {
		if len(pair) == 2 {
			formattedMetadata = append(formattedMetadata, fmt.Sprintf("%s|%s", pair[0], pair[1]))
		}
	}
	formattedMetadata = append(formattedMetadata, "created_at|DATETIME('now')")
	formattedMetadata = append(formattedMetadata, "updated_at|DATETIME('now')")

	formattedMetadataString := strings.Join(formattedMetadata, "\n")

	query := `Tengo un modelo de datos: ` + class + ` con los siguientes atributos y su tipo de dato correspondiente:
		` + formattedMetadataString + `
		Requiero construir basado en los datos anteriores las sentencias insert con data dummy, en total 5 sentencias para una base de datos sqlite, como las siguientes:
		-- DML statements [Dummy data]
		INSERT INTO products (name, description, price, created_at, updated_at)
			 VALUES ('Teléfono móvil', 'Smartphone de última generación', 799, DATETIME('now'), DATETIME('now'));
		INSERT INTO products (name, description, price, created_at, updated_at)
			 VALUES ('Camiseta', 'Camiseta de algodón', 20, DATETIME('now'), DATETIME('now'));
		INSERT INTO products (name, description, price, created_at, updated_at)
			 VALUES ('Sartén antiadherente', 'Sartén para cocinar', 35, DATETIME('now'), DATETIME('now'));
		INSERT INTO products (name, description, price, created_at, updated_at)
			 VALUES ('Balón de fútbol', 'Balón oficial de la FIFA', 50, DATETIME('now'), DATETIME('now'));
		INSERT INTO products (name, description, price, created_at, updated_at)
			 VALUES ('Muñeca', 'Muñeca de peluche para niños', 15, DATETIME('now'), DATETIME('now'));
		Es necesario no utilizar caracteres especiales ni comas en los posesivos en caso de ser información en inglés.
		Además considerar que la entidad para nombrar la tabla debe ser en plural en las sentencias insert.
		
		También requiero que generes a partir de los 2 primeros inserts la estructura de una request JSON. Es decir, 2 veces el siguiente ejemplo considerando el tipo de dato si son strings utilizar comillas, en caso de ser valores numéricos omitirlas.
		"name": "value",
		"description": "value",
		"price": 100000
		`

	resp, err := model.GenerateContent(
		ctx,
		genai.Text(query),
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Failed to marshal response: %v", err)
	}

	var data DDLData
	err = json.Unmarshal(respJSON, &data)
	if err != nil {
		log.Fatalf("Error al deserializar la respuesta: %v", err)
	}

	var parts []string
	for _, candidate := range data.Candidates {
		parts = append(parts, candidate.Content.Parts...)
	}

	return strings.Join(parts, "\n")
}

func ExtractInsertStatements(data string) string {
	re := regexp.MustCompile(`(?i)INSERT INTO [^\;]+;`)
	inserts := re.FindAllString(data, -1)
	return strings.Join(inserts, "\n")
}

func ExtractCreateUpdate(data string) (string, string) {
	re := regexp.MustCompile(`(?i)INSERT INTO ([^\(]+)\(([^\)]+)\)\s+VALUES\s+\(([^\)]+)\);`)
	matches := re.FindAllStringSubmatch(data, -1)

	if len(matches) < 2 {
		return "", ""
	}

	create := convertInsertToCurlBody(matches[0][2], matches[0][3])
	update := convertInsertToCurlBody(matches[1][2], matches[1][3])

	return create, update
}

func convertInsertToCurlBody(columns string, values string) string {
	columnList := strings.Split(columns, ",")
	valueList := strings.Split(values, ",")

	var result []string
	for i := range columnList {
		column := strings.TrimSpace(columnList[i])
		value := strings.TrimSpace(valueList[i])
		if column != "created_at" && column != "updated_at" {
			if isNumeric(value) {
				result = append(result, fmt.Sprintf("\"%s\": %s", column, value))
			} else {
				result = append(result, fmt.Sprintf("\"%s\": \"%s\"", column, value))
			}
		}
	}

	return strings.Join(result, ",\n")
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func AddDummyData(class string, classMetadata [][]string) DummyDataResult {
	dummyData := GenerateDummyData(class, classMetadata)
	insertStatements := ExtractInsertStatements(dummyData)
	createJSON, updateJSON := ExtractCreateUpdate(dummyData)

	return DummyDataResult{
		Inserts:    insertStatements,
		CreateJSON: createJSON,
		UpdateJSON: updateJSON,
	}
}
