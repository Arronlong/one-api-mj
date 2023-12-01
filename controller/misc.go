package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/model"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"version":             common.Version,
			"start_time":          common.StartTime,
			"email_verification":  common.EmailVerificationEnabled,
			"github_oauth":        common.GitHubOAuthEnabled,
			"github_client_id":    common.GitHubClientId,
			"system_name":         common.SystemName,
			"logo":                common.Logo,
			"footer_html":         common.Footer,
			"wechat_qrcode":       common.WeChatAccountQRCodeImageURL,
			"wechat_login":        common.WeChatAuthEnabled,
			"server_address":      common.ServerAddress,
			"pay_address":         common.PayAddress,
			"epay_id":             common.EpayId,
			"epay_key":            common.EpayKey,
			"price":               common.Price,
			"turnstile_check":     common.TurnstileCheckEnabled,
			"turnstile_site_key":  common.TurnstileSiteKey,
			"top_up_link":         common.TopUpLink,
			"chat_link":           common.ChatLink,
			"quota_per_unit":      common.QuotaPerUnit,
			"display_in_currency": common.DisplayInCurrencyEnabled,
		},
	})
	return
}

func GetNotice(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    common.OptionMap["Notice"],
	})
	return
}

func GetAbout(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    common.OptionMap["About"],
	})
	return
}

func GetMidjourney(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    common.OptionMap["Midjourney"],
	})
	return
}

func GetHomePageContent(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    common.OptionMap["HomePageContent"],
	})
	return
}

func SendEmailVerification(c *gin.Context) {
	email := c.Query("email")
	if err := common.Validate.Var(email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if common.EmailDomainRestrictionEnabled {
		allowed := false
		for _, domain := range common.EmailDomainWhitelist {
			if strings.HasSuffix(email, "@"+domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员启用了邮箱域名白名单，您的邮箱地址的域名不在白名单中",
			})
			return
		}
	}
	if model.IsEmailAlreadyTaken(email) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "邮箱地址已被占用",
		})
		return
	}
	code := common.GenerateVerificationCode(6)
	common.RegisterVerificationCodeWithKey(email, code, common.EmailVerificationPurpose)
	subject := fmt.Sprintf("%s邮箱验证邮件", common.SystemName)
	content := fmt.Sprintf("<p>您好，你正在进行%s邮箱验证。</p>"+
		"<p>您的验证码为: <strong>%s</strong></p>"+
		"<p>验证码 %d 分钟内有效，如果不是本人操作，请忽略。</p>", common.SystemName, code, common.VerificationValidMinutes)
	err := common.SendEmail(subject, email, content)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func SendPasswordResetEmail(c *gin.Context) {
	email := c.Query("email")
	if err := common.Validate.Var(email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if !model.IsEmailAlreadyTaken(email) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该邮箱地址未注册",
		})
		return
	}
	code := common.GenerateVerificationCode(0)
	common.RegisterVerificationCodeWithKey(email, code, common.PasswordResetPurpose)
	link := fmt.Sprintf("%s/user/reset?email=%s&token=%s", common.ServerAddress, email, code)
	subject := fmt.Sprintf("%s密码重置", common.SystemName)
	content := fmt.Sprintf("<p>您好，你正在进行%s密码重置。</p>"+
		"<p>点击 <a href='%s'>此处</a> 进行密码重置。</p>"+
		"<p>如果链接无法点击，请尝试点击下面的链接或将其复制到浏览器中打开：<br> %s </p>"+
		"<p>重置链接 %d 分钟内有效，如果不是本人操作，请忽略。</p>", common.SystemName, link, link, common.VerificationValidMinutes)
	err := common.SendEmail(subject, email, content)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

type PasswordResetRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func ResetPassword(c *gin.Context) {
	var req PasswordResetRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if req.Email == "" || req.Token == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	if !common.VerifyCodeWithKey(req.Email, req.Token, common.PasswordResetPurpose) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "重置链接非法或已过期",
		})
		return
	}
	password := common.GenerateVerificationCode(12)
	err = model.ResetUserPasswordByEmail(req.Email, password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	common.DeleteKey(req.Email, common.PasswordResetPurpose)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    password,
	})
	return
}

type AvailableModel struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
}

func GetAIChat_website(c *gin.Context) {
	var models []AvailableModel
	err := json.Unmarshal([]byte(common.AIChatModels), &models)
	if err != nil {
	  jsonStr := `[{"name": "gpt-3.5-turbo-16k","contentType": "Text","level": "NormalChat","levelId": 1}]`
		json.Unmarshal([]byte(jsonStr), &models)
	}

	c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"message": nil,
			"cnMessage": nil,
			"data": gin.H{
				"id": 1,
				"type": "Website",
				"typeId": 1,
				"content": "",
				"websiteContent": gin.H{
					"frontBaseUrl": "",
					"title": common.AIChatTitle,
					"mainTitle": common.AIChatMainTitle,
					"subTitle": common.AIChatSubTitle,
					"icp": nil,
					"globalJavaScript": nil,
					"loginPageSubTitle": "登录后可管理许可证",
					"registerPageSubTitle": "注册后赠送免费额度哦",
					"registerTypes": []string{
						"UsernameAndEmailWithCaptchaAndCode",
					},
					"registerEmailSuffix": "@",
					"pricingPageTitle": nil,
					"pricingPageSubTitle": nil,
					"chatPageSubTitle": nil,
					"sensitiveWordsTip": nil,
					"balanceNotEnough": nil,
					"hideGithubIcon": false,
					"botHello": nil,
					"logoUuid": nil,
					// "availableModels": []gin.H{
					// 	{
					// 		"name": "aichat智能助手",
					// 		"contentType": "Text",
					// 	},
					// },
					"availableModels": models,
					"registerForInviteCodeOnly": false,
					"auditAfterRegister": true,
					"auditingActions": []string{
						// "login",
						// "buyPackage",
					},
					"hideChatLogWhenNotLogin": false,
					"redeemCodePageTitle": "",
					"redeemCodePageSubTitle": "",
					"redeemCodePageBanner": "",
					"redeemCodePageTop": "",
					"redeemCodePageIndex": "",
					"redeemCodePageBottom": "",
					"defaultSystemTemplate": nil,
					"plugins": []gin.H{
						// {
						// 	"id": 1,
						// 	"uuid": "594a90be-0f21-4f72-9a22-4403e5028f81",
						// 	"name": "联网插件",
						// 	"logo": nil,
						// 	"alone": true,
						// 	"builtin": true,
						// 	"state": 10,
						// 	"createTime": "2023-08-27 16:19:09",
						// 	"updateTime": "2023-08-27 16:19:09",
						// },
					},
				},
				"emailContent": nil,
				"phoneContent": nil,
				"registerQuotaContent": nil,
				"sensitiveWordsContent": nil,
				"noticeContent": nil,
				"payContent": nil,
				"wechatContent": nil,
				"inviteRegisterContent": nil,
				"chatContent": nil,
				"drawContent": nil,
				"aboutContent": nil,
			},
		})
	return
}

func GetAIChat_notice(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": nil,
        "cnMessage": nil,
        "data": map[string]interface{}{
            "id":             1,
            "type":           "Notice",
            "typeId":         5,
            "content":        "",
            "websiteContent": nil,
            "emailContent":   nil,
            "phoneContent":   nil,
            "registerQuotaContent": nil,
            "sensitiveWordsContent": nil,
            "noticeContent": map[string]interface{}{
                "show":    common.AIChatNoticeShowEnabled,
                "splash":  common.AIChatNoticeSplashEnabled,
                "title":   common.AIChatNoticeTitle,
                "content": common.AIChatNoticeContent,
            },
            "payContent":            nil,
            "wechatContent":         nil,
            "inviteRegisterContent": nil,
            "chatContent":           nil,
            "drawContent":           nil,
        },
		})
	return
}
