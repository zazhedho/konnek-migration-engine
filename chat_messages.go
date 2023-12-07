package main

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"konnek-migration/models"
	"konnek-migration/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Create source DB connection
	scDB := utils.GetDBConnection()
	defer func(scDB *gorm.DB) {
		err := scDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection sourceDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(scDB)

	// Create destination DB connection
	dstDB := utils.GetDBNewConnection()
	defer func(dstDB *gorm.DB) {
		err := dstDB.Close()
		if err != nil {
			utils.WriteLog(fmt.Sprintf("Close Connection destinationsDb; ERROR: %+v", err), utils.LogLevelError)
		}
	}(dstDB)

	logID := uuid.NewV4()
	appName := "chat_messages"
	if os.Getenv("APP_NAME") != "" {
		appName = os.Getenv("APP_NAME")
	}
	logPrefix := fmt.Sprintf("[%v][%v]", logID, appName)
	utils.WriteLog(fmt.Sprintf("%s start...", logPrefix), utils.LogLevelDebug)

	tStart := time.Now()
	debug := 0
	debugT := time.Now()

	var dataChatMessages []models.ChatMessage
	// Get from file
	if os.Getenv("GET_FROM_FILE") != "" {
		utils.WriteLog(fmt.Sprintf("%s get from file %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelDebug)
		// Read the JSON file
		fileContent, err := ioutil.ReadFile("data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			fmt.Printf("%s Error reading file: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error reading file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}

		// Unmarshal the JSON data into the struct
		err = json.Unmarshal(fileContent, &dataChatMessages)
		if err != nil {
			fmt.Printf("%s Error unmarshalling: %v\n", logPrefix, err)
			utils.WriteLog(fmt.Sprintf("%s Error unmarshalling JSON: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
			return
		}
		debug++
		utils.WriteLog(fmt.Sprintf("%s [GET_FROM_FILE] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL_TIME: %s;", logPrefix, len(dataChatMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()

		err = os.Remove("data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch the data from existing PSQL database
		//Set the filters
		if os.Getenv("COMPANYID") != "" {
			scDB = scDB.Where("company_id = ?", os.Getenv("COMPANYID"))
		}

		if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") != "" {
			scDB = scDB.Where("created_at BETWEEN ? AND ?", os.Getenv("START_DATE"), os.Getenv("END_DATE"))
		} else if os.Getenv("START_DATE") != "" && os.Getenv("END_DATE") == "" {
			scDB = scDB.Where("created_at >=?", os.Getenv("START_DATE"))
		} else if os.Getenv("START_DATE") == "" && os.Getenv("END_DATE") != "" {
			scDB = scDB.Where("created_at <=?", os.Getenv("END_DATE"))
		}

		if os.Getenv("ORDER_BY") != "" {
			sortMap := map[string]string{
				"created_at": "created_at",
			}
			if strings.ToUpper(os.Getenv("ORDER_DIRECTION")) == "DESC" {
				scDB = scDB.Order(sortMap[os.Getenv("ORDER_BY")] + " DESC")
			} else {
				scDB = scDB.Order(sortMap[os.Getenv("ORDER_BY")])
			}
		}

		offset, _ := strconv.Atoi(os.Getenv("OFFSET"))
		limit, _ := strconv.Atoi(os.Getenv("LIMIT"))
		if offset > 0 {
			scDB = scDB.Offset(offset)
		}
		if limit > 0 {
			scDB = scDB.Limit(limit)
		}

		// query data dari source PSQL DB
		if err := scDB.Preload("ChatMedia").Find(&dataChatMessages).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		}
		totalFetch := len(dataChatMessages)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into the new PSQL database
	var errorMessages []models.ChatMessage
	var errorDuplicates []models.ChatMessage
	totalInserted := 0 //success insert
	for _, dataChatMessageEx := range dataChatMessages {
		messageType := "text"
		textMessage := ""
		if dataChatMessageEx.ChatMedia.ChatMessageId != uuid.Nil { // image, doc, voice, video
			messageType = getMediaType(dataChatMessageEx.ChatMedia.Media)
		} else {
			if messageType == "text" {
				textMessage = dataChatMessageEx.Message
			}
		}

		var payloadDecode map[string]interface{}
		var payloadLocation string
		var payloadTemplate string
		var payload string

		if err := json.Unmarshal([]byte(dataChatMessageEx.Message), &payloadDecode); err != nil {
			utils.WriteLog(fmt.Sprintf("%s; can't Unmarshal; type is 'text': %v", logPrefix, err), utils.LogLevelError)
		} else {
			if v, ok := payloadDecode["payload"].(map[string]interface{})["template_type"]; ok {
				payloadTemplate = v.(string)
				messageType = models.MessageTypeMap[payloadTemplate]
			} else if v, ok = payloadDecode["payload"].(map[string]interface{})["type"]; ok {
				payloadTemplate = v.(string)
				messageType = models.MessageTypeMap[payloadLocation]
			} else if _, ok = payloadDecode["payload"].(map[string]interface{})["key"]; ok {
				messageType = models.MessagePostback
			}

			var actions []models.Action
			if messageType == models.MessageButton {
				var buttonMessage models.ButtonMessage
				payloadAction := payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["actions"].([]map[string]interface{})
				for _, action := range payloadAction {
					actions = append(actions, models.Action{
						Key:   action["payload"].(map[string]interface{})["key"].(string),
						Title: action["label"].(string),
						Type:  action["type"].(string),
					})
				}
				buttonMessage.Body = actions
				buttonMessageByte, _ := json.Marshal(buttonMessage)
				payload = string(buttonMessageByte)
			} else if messageType == models.MessageList {
				var listMessage models.ListMessage

				listMessage.Body.Title = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["body"].(map[string]interface{})["text"].(string)
				listMessage.Header.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["header"].(map[string]interface{})["text"].(string)
				listMessage.Header.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["header"].(map[string]interface{})["type"].(string)
				listMessage.Footer.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["footer"].(map[string]interface{})["text"].(string)

				payloadSection := payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["action"].(map[string]interface{})["sections"].([]map[string]interface{})
				var listSection []models.Section
				for _, section := range payloadSection {
					actions = []models.Action{}
					for _, row := range section["rows"].([]map[string]interface{}) {
						actions = append(actions, models.Action{
							Description: row["description"].(string),
							Key:         row["id"].(string),
							Title:       row["title"].(string),
						})
					}
					listSection = append(listSection, models.Section{
						Actions: actions,
					})
				}
				listMessage.Body.Sections = listSection
				listMessageByte, _ := json.Marshal(listMessage)
				payload = string(listMessageByte)
			} else if messageType == models.MessagePostback {
				postbackMsg := models.PostbackMessage{
					Type:  payloadDecode["type"].(string),
					Title: payloadDecode["label"].(string),
					Key:   payloadDecode["payload"].(map[string]interface{})["key"].(string),
				}
				postbackMsgByte, _ := json.Marshal(postbackMsg)
				payload = string(postbackMsgByte)
			} else if messageType == models.MessageLocation {
				locationMsg := models.LocationMessage{
					Lat:        payloadDecode["payload"].(map[string]interface{})["latitude"].(float64),
					Lng:        payloadDecode["payload"].(map[string]interface{})["longitude"].(float64),
					Address:    payloadDecode["payload"].(map[string]interface{})["address"].(string),
					Name:       payloadDecode["type"].(string),
					LivePeriod: int(payloadDecode["payload"].(map[string]interface{})["live_period"].(float64)),
				}
				locationMsgByte, _ := json.Marshal(locationMsg)
				payload = string(locationMsgByte)
			} else if messageType == models.MessageCarousel {
				var carouselMessage models.CarouselMessage
				payloadItems := payloadDecode["payload"].(map[string]interface{})["items"].([]map[string]interface{})
				var listCarousel []models.CarouselStruct
				for _, item := range payloadItems {
					actions = []models.Action{}
					for _, action := range item["actions"].([]map[string]interface{}) {
						actions = append(actions, models.Action{
							Type:  action["type"].(string),
							Key:   action["payload"].(map[string]interface{})["key"].(string),
							Title: action["label"].(string),
						})
					}
					listCarousel = append(listCarousel, models.CarouselStruct{
						Title:       item["title"].(string),
						Description: item["text"].(string),
						MediaUrl:    item["thumbnailImageUrl"].(string),
						Actions:     actions,
					})
				}
				carouselMessage.Body.Carousel = listCarousel
				carouselMsgByte, _ := json.Marshal(carouselMessage)
				payload = string(carouselMsgByte)
			} else if dataChatMessageEx.FromType == 4 {
				messageType = models.MessageTemplate

				var templateMessage models.TemplateMessage
				templateMessage.Header.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["header"].(map[string]interface{})["text"].(string)
				templateMessage.Header.Type = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["header"].(map[string]interface{})["type"].(string)
				templateMessage.Footer.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["footer"].(map[string]interface{})["text"].(string)
				templateMessage.Body.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["body"].(map[string]interface{})["text"].(string)

				templateMessageByte, _ := json.Marshal(templateMessage)
				payload = string(templateMessageByte)
			}
		}

		mChatMessages := models.ChatMessages{
			Id:                dataChatMessageEx.Id,
			RoomId:            dataChatMessageEx.RoomId,
			SessionId:         dataChatMessageEx.SessionId,
			UserId:            dataChatMessageEx.UserId,
			MessageId:         dataChatMessageEx.MsgId,
			ProviderMessageId: dataChatMessageEx.TrxId,
			FromType:          strconv.Itoa(dataChatMessageEx.FromType),
			Type:              messageType,
			Text:              textMessage,
			Payload:           payload,
			Status:            dataChatMessageEx.Status,
			MessageTime:       &dataChatMessageEx.MessageTime,
			ReceivedAt:        &dataChatMessageEx.CreatedAt,
			ProcessedAt:       &dataChatMessageEx.MessageTime,
			DeleteTime:        dataChatMessageEx.DeletedAt,
			CreatedAt:         dataChatMessageEx.CreatedAt,
			CreatedBy:         uuid.Nil,
			UpdatedAt:         dataChatMessageEx.CreatedAt,
			UpdatedBy:         uuid.Nil,
			DeletedAt:         dataChatMessageEx.DeletedAt,
			DeletedBy:         uuid.Nil,
		}

		if err := dstDB.Create(&mChatMessages).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; insert error: %v", logPrefix, err), utils.LogLevelError)
			dataChatMessageEx.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, dataChatMessageEx)
					continue
				}
			}
			errorMessages = append(errorMessages, dataChatMessageEx)
			continue
		}
		totalInserted++
	}
	debug++
	utils.WriteLog(fmt.Sprintf("%s [PSQL] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

	//write error to file
	if len(errorMessages) > 0 {
		filename := fmt.Sprintf("%s_%s_%v", appName, time.Now().Format("2006_01_02"), time.Now().Unix())
		utils.WriteErrorMap(filename, errorMessages)
	}
	if len(errorDuplicates) > 0 {
		filename := fmt.Sprintf("%s_%s_%v_duplicate", appName, time.Now().Format("2006_01_02"), time.Now().Unix())
		utils.WriteErrorMap(filename, errorDuplicates)
	}

	utils.WriteLog(fmt.Sprintf("%s end; duration: %v", logPrefix, time.Now().Sub(tStart)), utils.LogLevelDebug)
}

func getMediaType(url string) string {
	fileExt := filepath.Ext(url)
	fileExt = strings.ToLower(fileExt)

	switch fileExt {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return "image"
	case ".mp4", ".avi", ".mkv", ".mov":
		return "video"
	case ".mp3", ".wav", ".ogg", ".flac", ".acc":
		return "audio"
	case ".doc", ".docx", ".xls", ".xlsx", ".pdf", ".txt", ".ppt", ".pptx", ".csv":
		return "document"
	default:
		return ""
	}
}
