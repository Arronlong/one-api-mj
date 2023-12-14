package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strconv"
	"time"
	"strings"
	"context"
  "regexp"
	
)

func UpdateMidjourneyTask() {
	//revocer
	imageModel := "midjourney"
	for {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("UpdateMidjourneyTask panic: %v", err)
			}
		}()
		time.Sleep(time.Duration(15) * time.Second)
		tasks := model.GetAllUnFinishTasks()
		if len(tasks) != 0 {
			log.Printf("检测到未完成的任务数有: %v", len(tasks))
			for _, task := range tasks {
				log.Printf("未完成的任务信息: %v", task)
				midjourneyChannel, err := model.GetChannelById(task.ChannelId, true)
				if err != nil {
					log.Printf("UpdateMidjourneyTask: %v", err)
					task.FailReason = fmt.Sprintf("获取渠道信息失败，请联系管理员，渠道ID：%d", task.ChannelId)
					task.Status = "FAILURE"
					task.Progress = "100%"
					err := task.Update()
					if err != nil {
						log.Printf("UpdateMidjourneyTask error: %v", err)
					}
					continue
				}
				requestUrl := fmt.Sprintf("%s/mj/task/%s/fetch", *midjourneyChannel.BaseURL, task.MjId)

			  switch midjourneyChannel.Type {

			  case common.ChannelTypeChatMj: // https://github.com/Licoy/ChatGPT-Midjourney/tree/v2
			    requestUrl = fmt.Sprintf("%s/mj/task/%s/fetch", *midjourneyChannel.BaseURL, task.MjId)

			  case common.ChannelTypeChatMjv3: // https://github.com/Licoy/ChatGPT-Midjourney
			    requestUrl = fmt.Sprintf("%s/task/status/%s", *midjourneyChannel.BaseURL, task.MjId)

			  case common.ChannelTypeMjProxy: //https://github.com/novicezk/midjourney-proxy
			    requestUrl = fmt.Sprintf("%s/mj/task/%s/fetch", *midjourneyChannel.BaseURL, task.MjId)

			  case common.ChannelTypeMjProxyPlus: //https://github.com/litter-coder/midjourney-proxy-plus
			    requestUrl = fmt.Sprintf("%s/mj/task/%s/fetch", *midjourneyChannel.BaseURL, task.MjId)

			  case common.ChannelTypeGoApiDraw: //https://docs.goapi.ai/docs/midjourney-api/midjourney-api-v2#inpaint
			    // requestUrl = fmt.Sprintf("%s/mj/v2/fetch", *midjourneyChannel.BaseURL, task.MjId)

			  case common.ChannelTypeAImageDraw: // https://jiao.nanjiren.online/t/topic/401
			    // requestUrl = fmt.Sprintf("%s/draw/info?uuid=%s", *midjourneyChannel.BaseURL, task.MjId)
			  default:
			    requestUrl = fmt.Sprintf("%s/mj/task/%s/fetch", *midjourneyChannel.BaseURL, task.MjId)

			  }

				log.Printf("requestUrl: %s", requestUrl)

				req, err := http.NewRequest("GET", requestUrl, bytes.NewBuffer([]byte("")))
				if err != nil {
					log.Printf("UpdateMidjourneyTask error: %v", err)
					continue
				}

				// 设置超时时间
				timeout := time.Second * 5
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				// 使用带有超时的 context 创建新的请求
				req = req.WithContext(ctx)

				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer midjourney-proxy")
				req.Header.Set("mj-api-secret", midjourneyChannel.Key)
				resp, err := httpClient.Do(req)
				if err != nil {
					log.Printf("UpdateMidjourneyTask error: %v", err)
					continue
				}
				defer resp.Body.Close()
				responseBody, err := io.ReadAll(resp.Body)
				log.Printf("responseBody: %s", string(responseBody))
				var responseItem Midjourney
				// err = json.NewDecoder(resp.Body).Decode(&responseItem)
				err = json.Unmarshal(responseBody, &responseItem)
				if err != nil {
					if strings.Contains(err.Error(), "cannot unmarshal number into Go struct field Midjourney.status of type string") {
		        var responseWithoutStatus MidjourneyWithoutStatus
		        var responseStatus MidjourneyStatus
		        err1 := json.Unmarshal(responseBody, &responseWithoutStatus)
		        err2 := json.Unmarshal(responseBody, &responseStatus)
		        if err1 == nil && err2 == nil {
							jsonData, err3 := json.Marshal(responseWithoutStatus)
							if err3 != nil {
								log.Fatalf("UpdateMidjourneyTask error1: %v", err3)
								continue
							}
							err4 := json.Unmarshal(jsonData, &responseStatus)
							if err4 != nil {
								log.Fatalf("UpdateMidjourneyTask error2: %v", err4)
								continue
							}
	            responseItem.Status = strconv.Itoa(responseStatus.Status)
		        } else {
	            log.Printf("UpdateMidjourneyTask error3: %v", err)
							continue
		        }
			    } else {
		        log.Printf("UpdateMidjourneyTask error4: %v", err)
						continue
			    }
				}
				
				if responseItem.SubmitTime == 0 {
					responseItem.SubmitTime = time.Now().UnixNano() / int64(time.Millisecond)
				}
				if responseItem.Status == "PROGRESS" {
					responseItem.Status = "IN_PROGRESS"
					if responseItem.StartTime == 0 {
						responseItem.StartTime = time.Now().UnixNano() / int64(time.Millisecond)
					}
				} else if responseItem.Status == "FAIL"{
					responseItem.Status = "FAILURE"
					responseItem.Progress = "100%"
				}

				if responseItem.Progress == "done"{
					responseItem.Progress = "100%"
					if responseItem.FinishTime == 0 {
						responseItem.FinishTime = time.Now().UnixNano() / int64(time.Millisecond)
					}
				}

				if midjourneyChannel.Type == common.ChannelTypeChatMjv3{
					responseItem.Description = "/imagine " + responseItem.Prompt
					responseItem.State = responseItem.MsgId
					responseItem.ImageUrl = responseItem.URI
				}

				if responseItem.Progress == "100%" && strings.HasPrefix(responseItem.ImageUrl, "http") {
					re := regexp.MustCompile(`_([a-zA-Z0-9-]+)\.png`)
			    match := re.FindStringSubmatch(responseItem.ImageUrl)
			    if len(match) > 1 {
			       responseItem.MsgHash = match[1]
			    }
				}

				task.Code = 1
				task.Progress = responseItem.Progress
				task.Prompt = responseItem.Prompt
				task.PromptEn = responseItem.PromptEn
				task.State = responseItem.State
				task.SubmitTime = responseItem.SubmitTime
				task.StartTime = responseItem.StartTime
				task.FinishTime = responseItem.FinishTime
				task.ImageUrl = responseItem.ImageUrl
				task.Status = responseItem.Status
				task.FailReason = responseItem.FailReason
				task.Description = responseItem.Description
				task.MsgHash = responseItem.MsgHash
				if task.Progress != "100%" && responseItem.FailReason != "" {
					log.Println(task.MjId + " 构建失败，" + task.FailReason)
					task.Progress = "100%"
					err = model.CacheUpdateUserQuota(task.UserId)
					if err != nil {
						log.Println("error update user quota cache: " + err.Error())
					} else {
						modelRatio := common.GetModelRatio(imageModel)
						groupRatio := common.GetGroupRatio("default")
						ratio := modelRatio * groupRatio
						quota := int(ratio * 1 * 1000)
						if quota != 0 {
							err := model.IncreaseUserQuota(task.UserId, quota)
							if err != nil {
								log.Println("fail to increase user quota")
							}
							logContent := fmt.Sprintf("%s 构图失败，补偿 %s", task.MjId, common.LogQuota(quota))
							model.RecordLog(task.UserId, 1, logContent)
						}
					}
				}

				err = task.Update()
				if err != nil {
					log.Printf("UpdateMidjourneyTask error5: %v", err)
				}
				log.Printf("UpdateMidjourneyTask success: %v", task)
			}
		}
	}
}

func GetAllMidjourney(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	logs := model.GetAllTasks(p*common.ItemsPerPage, common.ItemsPerPage)
	if logs == nil {
		logs = make([]*model.Midjourney, 0)
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
}

func GetUserMidjourney(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	userId := c.GetInt("id")
	log.Printf("userId = %d \n", userId)
	logs := model.GetAllUserTask(userId, p*common.ItemsPerPage, common.ItemsPerPage)
	if logs == nil {
		logs = make([]*model.Midjourney, 0)
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
}
