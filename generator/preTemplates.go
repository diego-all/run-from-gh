package generator

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

// Los que no son Req o Res van para la DB

var TypesVars = map[string]string{
	"handlers-typeEntityRequest":     "",
	"handlers-typeEntityResponse":    "",
	"handlers-varCreateEntityModels": "",
	"handlers-varGetEntResponse":     "",
	"handlers-varUpdateEntityModels": "",

	// EntityModels
	"models-typeEntityStruct":  "",
	"models-InsertStmt":        "",
	"models-InsertErr":         "",
	"models-GetOneQuery":       "",
	"models-GetOneErr":         "",
	"models-UpdateStmt":        "",
	"models-UpdateErr":         "",
	"models-GetAllQuery":       "",
	"models-GetAllErrRowsScan": "",
	"models-DeleteStmt":        "", // validar si realmente es necesario
}

var preTemplates = map[string]string{
	"cmd/api/handlers-entity-base.go": "/home/diegoall/MAESTRIA_ING/CLI/run-from-gh/base-templates/cmd/api/handlers-entity-base.txt",
	//"cmd/api/handlers-{{.Entity}}.go": "/home/diegoall/MAESTRIA_ING/CLI/run-from-gh/base-templates/cmd/api/handlers-entity.txt",
}

type PreTemplateData struct {
	Handlers_typeEntityRequest     string
	Handlers_typeEntityResponse    string
	Handlers_varCreateEntityModels string
	Handlers_varGetEntResponse     string
	Handlers_varUpdateEntityModels string
	Entity                         string
	LowerEntity                    string
}

// quizas sea generar Tipos o algo asi, todas las estructuras que dependen de la metadata de clases (atributos)
// structs or vars// {{.Entity}}Request  {{.Entity}}Response
// INTENTAR GENERANDO LOS TIPOS PRIMERO DE HANDLERS, LUEGO PARA MODELS U OTROS DE SER NECESARIO MANEJARLOS EN UN MAP

// func generateClassTags(class string, classMetadata map[string]string) (string) {
func generateClassTags(class string, classMetadata map[string]string) map[string]string {

	fmt.Println("Desde generateClassTags")

	fmt.Println("Class metadata", classMetadata)
	longitud := len(classMetadata)
	fmt.Println("longitud del map es:", longitud)
	fmt.Println("\n")

	var auxReqRes string
	var auxCreateEntModels string //Se uso para Update, solo se adicionan 2 campos
	var auxGetEntResponse string

	var reqResTypes []string
	var createEntModels []string
	var getEntResponse []string

	var multilineAuxReqResTypes string
	var multilineAuxCEntModels string
	var multilineAuxGEntResponse string

	var handlers_typeEntityRequest string
	var handlers_typeEntityResponse string
	var handlers_varCreateEntityModels string
	var handlers_varGetEntResponse string
	var handlers_varUpdateEntityModels string

	for attribute, value := range classMetadata {

		//fmt.Printf("Clave: %s, Valor: %s\n", attribute, value)
		fmt.Println("Capitalize alternativa nativa: ", strings.ToUpper(string(attribute[0]))+string(attribute[1:])) // toco esto para no usar mas dependencias.

		//auxReqRes = attribute + "\t" + value + "\t" + "`json:\"" + attribute + "\"`"
		auxReqRes = strings.ToUpper(string(attribute[0])) + string(attribute[1:]) + "\t" + value + "\t" + "`json:\"" + attribute + "\"`"
		auxCreateEntModels = strings.ToUpper(string(attribute[0])) + string(attribute[1:]) + ":\t" + "{{.LowerEntity}}Req." + strings.ToUpper(string(attribute[0])) + string(attribute[1:]) + ","
		auxGetEntResponse = strings.ToUpper(string(attribute[0])) + string(attribute[1:]) + ":\t" + "{{.LowerEntity}}." + strings.ToUpper(string(attribute[0])) + string(attribute[1:]) + ","

		//fmt.Println("auxReqRes", auxReqRes)
		//fmt.Println("auxCreateEntModels", auxCreateEntModels)
		//fmt.Println("AuxGetEntResponse", auxGetEntResponse)

		//Append de cada una de los atributos leidos
		reqResTypes = append(reqResTypes, auxReqRes)
		createEntModels = append(createEntModels, auxCreateEntModels)
		getEntResponse = append(getEntResponse, auxGetEntResponse)
	}

	// fmt.Println("Array de reqResTypes: ", reqResTypes)
	// fmt.Println("Array de createEntModels: ", createEntModels)
	// fmt.Println("Array de getEntResponse: ", getEntResponse)
	fmt.Println("\n")

	// Se verticalizan , creo que quedarian mejor con un while
	for i, _ := range reqResTypes {
		//fmt.Println("Valor de i", i, "Valor de j", j)
		multilineAuxReqResTypes = multilineAuxReqResTypes + reqResTypes[i] + "\n"
	}

	//fmt.Println("multilineAuxReqResTypes: \n ", multilineAuxReqResTypes+"\n")
	//fmt.Println("\n")

	handlers_typeEntityRequest = "type {{.LowerEntity}}Request struct {" + "\n" + multilineAuxReqResTypes + "}"
	//fmt.Println("\n")
	fmt.Println("handlers_typeEntityRequest: \n ", handlers_typeEntityRequest)
	fmt.Println("\n")

	handlers_typeEntityResponse = "type {{.LowerEntity}}Response struct {" + "\n" + multilineAuxReqResTypes + "}"
	fmt.Println("handlers_typeEntityResponse: \n ", handlers_typeEntityResponse)
	fmt.Println("\n")

	// Para createEntModels
	for i, _ := range createEntModels {
		//fmt.Println("Valor de i", i, "Valor de j", j)
		multilineAuxCEntModels = multilineAuxCEntModels + createEntModels[i] + "\n"
	}

	fmt.Println("multilineAuxCEntModels: \n ", multilineAuxCEntModels)
	fmt.Println("\n")
	// MINUSCULA
	//para create handlers_varCreateEntityModels
	handlers_varCreateEntityModels = "var {{.LowerEntity}} = models.{{.Entity}}{" + "\n" + multilineAuxCEntModels + "}"
	fmt.Println("\n")
	fmt.Println("handlers_varCreateEntityModels: \n ", handlers_varCreateEntityModels)

	// Para update handlers_varUpdateEntResponse
	handlers_varUpdateEntityModels = "var {{.LowerEntity}} = models.{{.Entity}}{" + "\n" + multilineAuxCEntModels + "UpdatedAt:   time.Now()," + "\n" + "Id:          {{.LowerEntity}}ID," + "\n" + "}"
	fmt.Println("\n")
	fmt.Println("handlers_varUpdateEntityModels: \n ", handlers_varUpdateEntityModels)

	for i, _ := range getEntResponse {
		//fmt.Println("Valor de i", i, "Valor de j", j)
		multilineAuxGEntResponse = multilineAuxGEntResponse + getEntResponse[i] + "\n"
	}

	//fmt.Println("multilineAuxGEntResponse: \n ", multilineAuxGEntResponse)
	fmt.Println("\n")

	handlers_varGetEntResponse = "var {{.LowerEntity}}Response = {{.LowerEntity}}Response{\n" + multilineAuxGEntResponse + "}"
	fmt.Println("\n")
	fmt.Println("handlers_varGetEntResponse: \n ", handlers_varGetEntResponse)

	fmt.Println("\n")
	// Generated Types and Vars
	TypesVars["handlers-typeEntityRequest"] = handlers_typeEntityRequest
	TypesVars["handlers-typeEntityResponse"] = handlers_typeEntityResponse
	TypesVars["handlers-varCreateEntityModels"] = handlers_varCreateEntityModels
	TypesVars["handlers-varGetEntResponse"] = handlers_varGetEntResponse
	TypesVars["handlers-varUpdateEntityModels"] = handlers_varUpdateEntityModels

	fmt.Println("TIPO FINAL: ", TypesVars)
	fmt.Println("\n")

	return TypesVars
	//return multiline // antes retornaba el primer type EntityRequest
}

// var generatedType = `var productResponse = {{.Entitrin}} productResponse{
// 	Name:        product.Name,
// 	Description: product.Description,
// 	Price:       product.Price,
// }`

func modifyBaseTemplates(preGeneratedTypes map[string]string) {

	fmt.Println("Desde modifyBaseTemplates")

	//var generatedType = "buenasnoches"

	// //Error al ejecutar la plantilla: template: fileContent:8:2: executing "fileContent" at <.handlers_typeEntityRequest>: handlers_typeEntityRequest is an unexported field of struct type generator.preTemplateData
	preData := PreTemplateData{
		Handlers_typeEntityRequest:     preGeneratedTypes["handlers-typeEntityRequest"],
		Handlers_typeEntityResponse:    preGeneratedTypes["handlers-typeEntityResponse"],
		Handlers_varCreateEntityModels: preGeneratedTypes["handlers-varCreateEntityModels"],
		Handlers_varGetEntResponse:     preGeneratedTypes["handlers-varGetEntResponse"],
		Handlers_varUpdateEntityModels: preGeneratedTypes["handlers-varUpdateEntityModels"],
		// quizas pueda ser {{.UpperEntity}}
		Entity: "{{.Entity}}",
		//entity:  "{{.entity}}",   // NO funciona con minusculas seguir indagando
		LowerEntity: "{{.LowerEntity}}",
	}

	fmt.Println(preData)

	// 	for attribute, value := range preGeneratedTypes {
	for projectFile, templatePath := range preTemplates {

		//fmt.Printf("Clave: %s, Valor: %s\n", projectFile, templatePath)
		fmt.Println("Path y Content es: ", projectFile, templatePath)

		// Si hay contenido de plantilla, procesarlo
		if templatePath != "" {

			content, err := ioutil.ReadFile(templatePath)
			if err != nil {
				fmt.Println("Error al leer la plantilla:", err)
				continue
			}

			fmt.Println("CONTENT:", string(content))

			tmpl, err := template.New("fileContent").Parse(string(content))
			fmt.Println("tmpl es:", tmpl)
			if err != nil {
				fmt.Println("Error al parsear la plantilla:", err)
				continue
			}

			// Crear o abrir el archivo de destino en modo escritura
			file, err := os.Create(templatePath)
			//file, err := os.Create(projectFile)
			if err != nil {
				fmt.Println("Error al crear el archivo:", err)
				continue
			}
			defer file.Close()

			if err := tmpl.Execute(file, preData); err != nil {
				fmt.Println("Error al ejecutar la plantilla:", err)
				continue
			}
			fmt.Println("Archivo procesado correctamente:", projectFile)

		}

	}

	fmt.Println("\n")

}

// Generate Create table
func generateDDLStatement(class string, classMetadata map[string]string) string {

	fmt.Println("Desde generateDDLStatement", class)

	fmt.Println("Class metadata", classMetadata)
	longitud := len(classMetadata)
	fmt.Println("longitud del map es:", longitud)
	fmt.Println("\n")

	for attribute, value := range classMetadata {
		fmt.Printf("Clave: %s, Valor: %s\n", attribute, value)

	}

	return ""
}

func generateEntityModels() {

}
