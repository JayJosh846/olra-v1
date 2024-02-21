package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"olra-v1/internal/database"
	"olra-v1/internal/structs"
	"olra-v1/middleware"
	"olra-v1/services"
	helpers "olra-v1/utils"
)

func CreateGroup(c *gin.Context) {
	// Initialize context and defer cancellation
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Retrieve user information from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "User not found",
			"data":          "",
		})
		return
	}

	// Check if user information is in a valid format
	userStruct, ok := user.(middleware.User)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User not a valid struct",
			"data":          "",
		})
		return
	}

	// Parse group request from JSON body
	var groupRequest structs.GroupRequest
	if err := c.BindJSON(&groupRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	// Validate group request
	validationErr := Validate.Struct(groupRequest)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       validationErr.Error(),
			"data":          "",
		})
		return
	}

	for _, item := range groupRequest.GroupMembers {
		validationErr := Validate.Struct(item)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         true,
				"response code": 400,
				"message":       validationErr.Error(),
				"data":          "",
			})
			return
		}
	}

	validationError := helpers.ValidateTagRequest(groupRequest.Tag)
	if validationError != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       validationError.Error(),
			"data":          "",
		})
		return
	}

	// Check if group tag name already exists
	var existingGroup database.Group
	if err := database.DB.Where("group_tag = ?", groupRequest.Tag).First(&existingGroup).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Group tag name already exists. Please use another name.",
			"data":          "",
		})
		return
	}

	// Iterate over group members and perform necessary checks
	for i := range groupRequest.GroupMembers {
		var friend database.User
		if err := database.DB.Where("tag = ?", groupRequest.GroupMembers[i].Friend).First(&friend).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         true,
				"response code": 400,
				"message":       "Friend with the tag does not exist.",
				"data":          "",
			})
			return
		}

		// Check if the user is trying to add themselves to the group
		if friend.UserID == *userStruct.UserId {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         true,
				"response code": 400,
				"message":       "You cannot add yourself to the group as a group creator",
				"data":          "",
			})
			return
		}
	}
	// Create the group
	group := database.Group{
		GroupName:   groupRequest.GroupName,
		GroupTag:    groupRequest.Tag,
		CreatedBy:   *userStruct.UserId,
		GroupBudget: groupRequest.Amount,
		AdminID:     *userStruct.UserId,
	}
	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to create group",
			"data":          "",
		})
		return
	}

	// Generate a random account number for the group's virtual wallet
	generatedAccountNumber := helpers.GenerateRandomAccountNumber()

	// Create a virtual wallet for the group
	groupVirtualAccount := database.GroupVirtualAccount{
		GroupVirtualAccountBank:   "Guaranty Trust Bank",
		GroupVirtualAccountNumber: generatedAccountNumber,
		GroupVirtualAccountName:   groupRequest.GroupName,
		GroupID:                   group.GroupID,
	}
	if err := database.DB.Create(&groupVirtualAccount).Error; err != nil {
		database.DB.Delete(&group)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to create virtual account",
			"data":          "",
		})
		return
	}

	// Iterate over group members and create group members
	for _, friendTag := range groupRequest.GroupMembers {
		var friend database.User
		database.DB.Where("tag = ?", friendTag.Friend).First(&friend)
		groupMember := database.GroupMember{
			GroupID:  group.GroupID,
			UserID:   friend.UserID,
			JoinedAt: time.Now(),
		}
		if err := database.DB.Create(&groupMember).Error; err != nil {
			database.DB.Delete(&groupVirtualAccount)
			database.DB.Delete(&group)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":         true,
				"response code": 500,
				"message":       "Failed to group member",
				"data":          "",
			})
			return
		}
	}

	// Respond with success message and created group data
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Group created successfully",
		"data":          group,
	})
}

func SendGroupFunds(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	groupId := c.Query("groupId")
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "User not found",
			"data":          "",
		})
		return
	}
	userStruct, ok := user.(middleware.User)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User not a valid struct",
			"data":          "",
		})
		return
	}
	var (
		sendGroupFundsRequest structs.SendGroupFundsRequest
		groupAccount          database.GroupVirtualAccount
		existingUser          database.User
		group                 database.Group
		userAccount           database.VirtualAccount
		// groupMember           database.GroupMember
		fundsResponse structs.FundsResponse
	)
	if err := c.BindJSON(&sendGroupFundsRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	validationErr := Validate.Struct(sendGroupFundsRequest)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       validationErr.Error(),
			"data":          "",
		})
		return
	}
	// Check for group wallet
	if err := database.DB.Where(
		"group_id = ?", groupId,
	).First(&groupAccount).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":         true,
			"response code": 404,
			"message":       "Could not find group wallet.",
			"data":          "",
		})
		return
	}
	if err := database.DB.Where(
		"group_id = ?", groupId,
	).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":         true,
			"response code": 404,
			"message":       "Could not find group.",
			"data":          "",
		})
		return
	}
	// Check user details
	if err := database.DB.Where(
		"user_id = ?", userStruct.UserId,
	).First(&existingUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":         true,
			"response code": 404,
			"message":       "User does not exist.",
			"data":          "",
		})
		return
	}
	// Check for user's bank account details
	if err := database.DB.Where(
		"user_id = ?", userStruct.UserId,
	).First(&userAccount).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":         true,
			"response code": 404,
			"message":       "User wallet does not exist.",
			"data":          "",
		})
		return
	}
	// Check if user belongs to group
	// if err := database.DB.Where(
	// 	"user_id = ?", userStruct.UserId,
	// ).First(&groupMember).Error; err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"error":         true,
	// 		"response code": 404,
	// 		"message":       "User does not belong to this group",
	// 		"data":          "",
	// 	})
	// 	return
	// }
	if userAccount.Balance < sendGroupFundsRequest.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Insufficient balance to send",
			"data":          "",
		})
		return
	}
	sentGroupBalance := sendGroupFundsRequest.Amount + groupAccount.Balance
	if err := database.DB.Model(&groupAccount).Update(
		"balance", sentGroupBalance,
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to update group account balance",
			"data":          "",
		})
		return
	}
	debitedBalance := userAccount.Balance - sendGroupFundsRequest.Amount
	if err := database.DB.Model(&userAccount).Update(
		"balance", debitedBalance,
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to update user account balance",
			"data":          "",
		})
		return
	}

	transaction := database.Transaction{
		UserID:             *userStruct.UserId,
		TransactionEnviron: "withinOlra",
		TransactionType:    "group-payment",
		Amount:             sendGroupFundsRequest.Amount,
		Description:        sendGroupFundsRequest.Description,
		Receiver:           group.GroupTag,
		Sender:             existingUser.Tag,
		Status:             "completed",
		TransactionDate:    time.Now(),
	}
	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to create transaction instance",
			"data":          "",
		})
		return
	}
	smsFundsResponse, errr := services.DebitGroupFundsSMS(
		existingUser.PhoneNumber,
		group.GroupName,
		sendGroupFundsRequest.Amount,
		userAccount.Balance,
	)
	if errr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Could not send debit funds sms",
			"data":          "",
		})
		return
	}
	ee := json.Unmarshal([]byte(smsFundsResponse), &fundsResponse)
	if ee != nil {
		log.Println("Error:", ee)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Group transaction successful",
		"data":          fundsResponse,
	})
}

func GroupRoutes(rg *gin.RouterGroup) {
	grouproute := rg.Group("/group")
	grouproute.POST(
		"/create-group",
		middleware.AuthMiddleware,
		CreateGroup,
	)
	grouproute.POST(
		"/group-transfer",
		middleware.AuthMiddleware,
		SendGroupFunds,
	)

}
