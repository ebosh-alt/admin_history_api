package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"net/http"
	"strconv"
	"time"
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

	if v := c.Query("page"); v != "" {
		if n, _ := strconv.ParseInt(v, 10, 32); n > 0 {
			req.Page = int32(n)
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, _ := strconv.ParseInt(v, 10, 32); n > 0 {
			req.Limit = int32(n)
		}
	}
	if v := c.Query("status"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			req.Status = wrapperspb.Bool(b)
		}
	}
	if v := c.Query("accepted_offer"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			req.AcceptedOffer = wrapperspb.Bool(b)
		}
	}
	if v := c.Query("date_from"); v != "" {
		if sec, _ := strconv.ParseInt(v, 10, 64); sec > 0 {
			req.DateFrom = timestamppb.New(time.Unix(sec, 0).UTC())
		}
	}
	if v := c.Query("date_to"); v != "" {
		if sec, _ := strconv.ParseInt(v, 10, 64); sec > 0 {
			req.DateTo = timestamppb.New(time.Unix(sec, 0).UTC())
		}
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	resp, err := s.Usecase.UsersList(c, req)
	if err != nil {
		s.log.Error("failed to get users list", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректные данные"})
		return
	}

	m := protojson.MarshalOptions{EmitUnpopulated: true}
	b, _ := m.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
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
