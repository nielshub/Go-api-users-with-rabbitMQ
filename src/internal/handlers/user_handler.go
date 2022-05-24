package handlers

import (
	"Faceit/src/internal/model"
	"Faceit/src/internal/ports"
	"Faceit/src/log"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xeipuuv/gojsonschema"
)

type UserHandler struct {
	router      *gin.RouterGroup
	userService ports.UserService
}

func NewUserHandler(app *gin.RouterGroup, userService ports.UserService) UserHandler {
	userAPI := UserHandler{userService: userService}

	userRooter := app.Group("/user")
	userRooter.POST("/create", userAPI.userCreate)
	userRooter.PUT("/update/:userId", userAPI.userUpdate)
	userRooter.POST("/delete/:userId", userAPI.userDelete)
	userRooter.GET("/get", userAPI.getUsers)

	userAPI.router = userRooter

	return userAPI
}

func (uh *UserHandler) userCreate(c *gin.Context) {
	var newUser model.User
	newUserBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Logger.Error().Msgf("Could not read the body. Error: %s", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Could not read the body"})
		return
	}
	err = json.Unmarshal(newUserBody, &newUser)
	if err != nil {
		log.Logger.Error().Msgf("Could not unmarshal the json body. Error: %s", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Could not unmarshal the json body"})
		return
	}

	if err = jsonSchemaCheck(newUser); err != nil {
		log.Logger.Error().Msgf("Wrong body struct. Does not match with jsonSchema. Error: %s", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Wrong body struct. Does not match with jsonSchema."})
		return
	}

	newUser.InitializeTime()

	createdUser, err := uh.userService.CreateUser(c.Request.Context(), newUser)
	if err != nil {
		log.Logger.Error().Msgf("Error creating user. Error: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error creating user"})
		return
	}
	response := "User has been created properly. User ID: " + createdUser.ID

	c.JSON(http.StatusOK, response)

}

func (uh *UserHandler) userUpdate(c *gin.Context) {
	id := c.Param("userId")
	var updatedUser model.User
	updatedUserBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Could not read the body"})
		return
	}
	err = json.Unmarshal(updatedUserBody, &updatedUser)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Could not unmarshal the json body"})
		return
	}

	if err = jsonSchemaCheck(updatedUser); err != nil {
		log.Logger.Error().Msgf("Wrong body struct. Does not match with jsonSchema. Error: %s", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Wrong body struct. Does not match with jsonSchema."})
		return
	}

	finalUser, err := uh.userService.UpdateUser(c.Request.Context(), id, updatedUser)
	if err != nil {
		log.Logger.Error().Msgf("Error updating user. Error: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error updating user"})
		return
	}
	response := "User:" + finalUser.ID + " has been updated properly."

	c.JSON(http.StatusOK, response)

}

func (uh *UserHandler) userDelete(c *gin.Context) {
	id := c.Param("userId")

	err := uh.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		log.Logger.Error().Msgf("Error deleting user. Error: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error deleting user"})
		return
	}
	response := "User:" + id + " has been deleted properly."

	c.JSON(http.StatusOK, response)

}

func (uh *UserHandler) getUsers(c *gin.Context) {
	key := c.Query("key")
	value := c.Query("value")

	users, err := uh.userService.GetUsers(c.Request.Context(), key, value)
	if err != nil {
		log.Logger.Error().Msgf("Error getting users. Error: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error getting users"})
		return
	}
	response := users

	c.JSON(http.StatusOK, response)

}

func jsonSchemaCheck(user model.User) error {
	loader := gojsonschema.NewReferenceLoader(os.Getenv("JSONSCHEMAPATH"))
	result, err := gojsonschema.Validate(loader, gojsonschema.NewGoLoader(user))
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	} else {
		// Print out each error
		for _, desc := range result.Errors() {
			log.Logger.Error().Msgf("- %s\n", desc)
		}
		return errors.New("the body of the request is not valid")
	}
}
