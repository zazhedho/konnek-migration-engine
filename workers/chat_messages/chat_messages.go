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
	utils.Init()

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

	startDate := os.Getenv("START_DATE")
	endDate := os.Getenv("END_DATE")
	limit, _ := strconv.Atoi(os.Getenv("LIMIT"))

reFetch:

	scDB = utils.GetDBConnection()
	//Set the filters
	if os.Getenv("COMPANYID") != "" {
		scDB = scDB.Joins("JOIN room_details ON chat_messages.room_id = room_details.id").Where("room_details.company_id = ?", os.Getenv("COMPANYID"))
	}

	if startDate != "" && endDate != "" {
		scDB = scDB.Where("chat_messages.created_at BETWEEN ? AND ?", startDate, endDate)
	} else if startDate != "" && endDate == "" {
		scDB = scDB.Where("chat_messages.created_at >=?", startDate)
	} else if startDate == "" && endDate != "" {
		scDB = scDB.Where("chat_messages.created_at <=?", endDate)
	}

	if os.Getenv("ORDER_BY") != "" {
		sortMap := map[string]string{
			os.Getenv("ORDER_BY"): "chat_messages." + os.Getenv("ORDER_BY"),
		}
		if strings.ToUpper(os.Getenv("ORDER_DIRECTION")) == "DESC" {
			scDB = scDB.Order(sortMap[os.Getenv("ORDER_BY")] + " DESC")
		} else {
			scDB = scDB.Order(sortMap[os.Getenv("ORDER_BY")])
		}
	}

	offset, _ := strconv.Atoi(os.Getenv("OFFSET"))
	if offset > 0 {
		scDB = scDB.Offset(offset)
	}
	if limit > 0 {
		scDB = scDB.Limit(limit)
	}

	totalFetch := 0

	// Get from file
	if os.Getenv("GET_FROM_FILE") != "" {
		utils.WriteLog(fmt.Sprintf("%s get from file %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelDebug)
		// Read the JSON file
		fileContent, err := ioutil.ReadFile("../../data/" + os.Getenv("GET_FROM_FILE"))
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

		err = os.Remove("../../data/" + os.Getenv("GET_FROM_FILE"))
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s Error Delete file: %s", logPrefix, os.Getenv("GET_FROM_FILE")), utils.LogLevelError)
		}
	} else {
		//Fetch the data from existing PSQL database

		// query data dari source PSQL DB
		if err := scDB.Preload("ChatMedia").Find(&dataChatMessages).Error; err != nil {
			utils.WriteLog(fmt.Sprintf("%s; fetch error: %v", logPrefix, err), utils.LogLevelError)
		}
		totalFetch = len(dataChatMessages)

		debug++
		utils.WriteLog(fmt.Sprintf("%s [FETCH] [>= '%v' LIMIT: %v] TOTAL_FETCH: %d DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, startDate, limit, totalFetch, debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)
		debugT = time.Now()
	}

	//Insert into the new PSQL database
	var errorMessages []models.ChatMessage
	var errorDuplicates []models.ChatMessage
	totalInserted := 0 //success insert
	for i, dataChatMessageEx := range dataChatMessages {
		var payload string
		messageType := "text"
		textMessage := ""
		if dataChatMessageEx.ChatMedia.ChatMessageId != uuid.Nil { // image, doc, voice, video
			messageType = getMediaType(dataChatMessageEx.ChatMedia.Media)
			splitMessage := strings.Split(dataChatMessageEx.Message, " - ")
			lengthMessage := len(splitMessage) - 1
			size, _ := strconv.Atoi(strings.Trim(strings.TrimSpace(splitMessage[lengthMessage]), " KB"))
			mediaMessage := models.MediaMessage{
				Id:   dataChatMessageEx.ChatMedia.Id.String(),
				Url:  dataChatMessageEx.ChatMedia.Media,
				Name: dataChatMessageEx.Message,
				Size: size,
			}
			imageMessageByte, _ := json.Marshal(mediaMessage)
			payload = string(imageMessageByte)
			textMessage = dataChatMessageEx.Message
		} else {
			if messageType == "text" {
				textMessage = dataChatMessageEx.Message
			}
		}

		var payloadDecode map[string]interface{}
		var payloadValDecode map[string]interface{}
		var payloadLocation string
		var payloadTemplate string

		if err := json.Unmarshal([]byte(dataChatMessageEx.Message), &payloadDecode); err != nil {
			utils.WriteLog(fmt.Sprintf("%s; can't Unmarshal; type is 'text': %v", logPrefix, err), utils.LogLevelError)
		} else {
			if v, ok := payloadDecode["payload"]; ok {
				switch v.(type) {
				case map[string]interface{}:
					if v, ok = payloadDecode["payload"].(map[string]interface{})["template_type"]; ok {
						payloadTemplate = v.(string)
						if dataChatMessageEx.FromType != 4 {
							messageType = models.MessageTypeMap[payloadTemplate]
						}
					}
				case string:
					if err = json.Unmarshal([]byte(payloadDecode["payload"].(string)), &payloadValDecode); err != nil {
						utils.WriteLog(fmt.Sprintf("%s; can't Unmarshal; type is 'text': %v", logPrefix, err), utils.LogLevelError)
					}
					if v, ok = payloadValDecode["type"]; ok {
						payloadLocation = v.(string)
						if dataChatMessageEx.FromType != 4 {
							messageType = models.MessageTypeMap[payloadLocation]
						}
					} else if _, ok = payloadValDecode["key"]; ok {
						messageType = models.MessagePostback
					}
				}
			}

			var actions []models.Action
			if messageType == models.MessageButton {
				var buttonMessage models.ButtonMessage
				payloadAction := payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["actions"].([]interface{})
				for _, action := range payloadAction {
					actions = append(actions, models.Action{
						Key:   action.(map[string]interface{})["payload"].(map[string]interface{})["key"].(string),
						Title: action.(map[string]interface{})["label"].(string),
						Type:  action.(map[string]interface{})["type"].(string),
					})
				}
				buttonMessage.Body = actions
				buttonMessage.Header.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["text"].(string)
				buttonMessage.Header.Type = "text"
				buttonMessageByte, _ := json.Marshal(buttonMessage)
				payload = string(buttonMessageByte)
				textMessage = buttonMessage.Header.Text
			} else if messageType == models.MessageList {
				var listMessage models.ListMessage

				listMessage.Body.Title = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["body"].(map[string]interface{})["text"].(string)
				listMessage.Header.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["header"].(map[string]interface{})["text"].(string)
				listMessage.Header.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["header"].(map[string]interface{})["type"].(string)
				listMessage.Footer.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["footer"].(map[string]interface{})["text"].(string)

				payloadSection := payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["action"].(map[string]interface{})["sections"].([]interface{})
				var listSection []models.Section
				for _, section := range payloadSection {
					actions = []models.Action{}
					for _, row := range section.(map[string]interface{})["rows"].([]interface{}) {
						actions = append(actions, models.Action{
							Description: row.(map[string]interface{})["description"].(string),
							Key:         row.(map[string]interface{})["id"].(string),
							Title:       row.(map[string]interface{})["title"].(string),
						})
					}
					listSection = append(listSection, models.Section{
						Actions: actions,
					})
				}
				listMessage.Body.Sections = listSection
				listMessageByte, _ := json.Marshal(listMessage)
				payload = string(listMessageByte)
				textMessage = listMessage.Body.Title
			} else if messageType == models.MessagePostback {
				postbackMsg := models.PostbackMessage{
					Type:  payloadDecode["type"].(string),
					Title: payloadDecode["label"].(string),
					Key:   payloadValDecode["key"].(string),
					Value: payloadDecode["label"].(string),
				}
				postbackMsgByte, _ := json.Marshal(postbackMsg)
				payload = string(postbackMsgByte)
				textMessage = postbackMsg.Value
			} else if messageType == models.MessageLocation {
				livePeriod := 0
				if val, isset := payloadValDecode["live_period"]; isset {
					livePeriod = int(val.(float64))
				}
				locationMsg := models.LocationMessage{
					Lat:        payloadValDecode["latitude"].(float64),
					Lng:        payloadValDecode["longitude"].(float64),
					Address:    payloadValDecode["address"].(string),
					Name:       payloadDecode["label"].(string),
					LivePeriod: livePeriod,
				}
				locationMsgByte, _ := json.Marshal(locationMsg)
				payload = string(locationMsgByte)
				textMessage = locationMsg.Name
			} else if messageType == models.MessageCarousel {
				var carouselMessage models.CarouselMessage
				payloadItems := payloadDecode["payload"].(map[string]interface{})["items"].([]interface{})
				var bodyCarousel []models.BodyCarousel
				for _, item := range payloadItems {
					actions = []models.Action{}
					for _, action := range item.(map[string]interface{})["actions"].([]interface{}) {
						actions = append(actions, models.Action{
							Type:  action.(map[string]interface{})["type"].(string),
							Key:   action.(map[string]interface{})["payload"].(map[string]interface{})["key"].(string),
							Title: action.(map[string]interface{})["label"].(string),
						})
					}
					bodyCarousel = append(bodyCarousel, models.BodyCarousel{
						Title:       item.(map[string]interface{})["title"].(string),
						Description: item.(map[string]interface{})["text"].(string),
						MediaUrl:    item.(map[string]interface{})["thumbnailImageUrl"].(string),
						Actions:     actions,
					})
				}
				carouselMessage.Body = bodyCarousel
				carouselMsgByte, _ := json.Marshal(carouselMessage)
				payload = string(carouselMsgByte)
				textMessage = ""
			} else if dataChatMessageEx.FromType == 4 {
				messageType = models.MessageTemplate

				var templateMessage models.TemplateMessage
				templateMessage.Header.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["header"].(map[string]interface{})["text"].(string)
				templateMessage.Header.Type = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["header"].(map[string]interface{})["type"].(string)
				templateMessage.Footer.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["footer"].(map[string]interface{})["text"].(string)
				templateMessage.Body.Text = payloadDecode["payload"].(map[string]interface{})["items"].(map[string]interface{})["body"].(map[string]interface{})["text"].(string)

				templateMessageByte, _ := json.Marshal(templateMessage)
				payload = string(templateMessageByte)
				textMessage = templateMessage.Body.Text
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

		err := dstDB.Create(&mChatMessages).Error
		if err != nil {
			utils.WriteLog(fmt.Sprintf("%s; insert error: %v", logPrefix, err), utils.LogLevelError)
			dataChatMessageEx.Error = err.Error()
			if errCode, ok := err.(*pq.Error); ok {
				if errCode.Code == "23505" { //unique_violation
					errorDuplicates = append(errorDuplicates, dataChatMessageEx)
				}
			}
			errorMessages = append(errorMessages, dataChatMessageEx)
		}
		totalInserted++

		if i >= limit {
			debug++
			utils.WriteLog(fmt.Sprintf("%s [PSQL] [>= '%v' LIMIT: %v] TOTAL_FETCH: %d; TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, startDate, limit, totalFetch, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

			startDate = dataChatMessageEx.CreatedAt.Format("2006-01-02 15:04:05.999999999+07")
			utils.WriteLog(fmt.Sprintf("last created_at:%v; set startDate:%v\n", dataChatMessageEx.CreatedAt, startDate), utils.LogLevelDebug)

			goto reFetch
		}
	}
	//debug++
	//utils.WriteLog(fmt.Sprintf("%s [PSQL] TOTAL_INSERTED: %d; TOTAL_ERROR: %v DEBUG: %d; TIME: %s; TOTAL TIME EXECUTION: %s;", logPrefix, totalInserted, len(errorMessages), debug, time.Now().Sub(debugT), time.Now().Sub(tStart)), utils.LogLevelDebug)

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
