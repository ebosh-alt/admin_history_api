package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
)

func (s *Server) GetUser(c *gin.Context) {
	req := &protos.UserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("failed to get user: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userProto, err := s.Usecase.GetUser(c, req)
	if err != nil {
		s.log.Error("failed to get user: "+err.Error(), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, userProto)

}

func (s *Server) UsersList(c *gin.Context) {
	req := &protos.UsersListRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("failed to get users: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userListProto, err := s.Usecase.UsersList(c, req)
	if err != nil {
		s.log.Error("failed to get users list: "+err.Error(), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, userListProto)
}

func (s *Server) UpdateUser(c *gin.Context) {
	req := &protos.UpdateUserRequest{}
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		s.log.Error("failed to read request: "+err.Error(), zap.String("request", string(data)))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}
	if err := protojson.Unmarshal(data, req); err != nil {
		s.log.Error("failed to unmarshal request: "+err.Error(), zap.String("request", string(data)))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос(типы данных или их формат)"})
		return
	}

	status, err := s.Usecase.UpdateUser(c, req)
	if err != nil {
		s.log.Error("failed to update user: "+err.Error(), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, status)
}
