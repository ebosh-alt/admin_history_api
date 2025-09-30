package server

import (
	"admin_history/config"
	"admin_history/internal/delivery/http/middleware"
	"admin_history/internal/usecase"
	protos "admin_history/pkg/proto/gen/go"
	"context"
	_ "fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
)

type Server struct {
	log        *zap.Logger
	cfg        *config.Config
	serv       *gin.Engine
	Usecase    usecase.InterfaceUsecase
	middleware *middleware.Middleware
}

var (
	unmarshalJSON = protojson.UnmarshalOptions{DiscardUnknown: true}
	marshalJSON   = protojson.MarshalOptions{EmitUnpopulated: true}
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

func (s *Server) GetQuestionnaire(c *gin.Context) {
	req := &protos.QuestionnaireRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("failed to get questionnaire: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}

	qProto, err := s.Usecase.GetQuestionnaire(c, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, qProto)
}
func (s *Server) QuestionnairesList(c *gin.Context) {
	req := &protos.QuestionnairesListRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("failed to get questionnaires: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}

	qProto, err := s.Usecase.GetQuestionnairesList(c, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, qProto)

}

func (s *Server) UpdateQuestionnaire(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req protos.UpdateQuestionnaireRequest
	if err := unmarshalJSON.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
		return
	}
	resp, err := s.Usecase.UpdateQuestionnaire(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

func (s *Server) GetChat(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) ChatsList(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (s *Server) GetStatistics(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func NewServer(logger *zap.Logger, cfg *config.Config, uc usecase.InterfaceUsecase, middleware *middleware.Middleware) (*Server, error) {
	return &Server{
		log:        logger,
		cfg:        cfg,
		serv:       gin.Default(),
		Usecase:    uc,
		middleware: middleware,
	}, nil
}

func (s *Server) OnStart(_ context.Context) error {
	s.CreateController()
	go func() {
		s.log.Debug("server started")
		if err := s.serv.Run(s.cfg.Server.Host + ":" + s.cfg.Server.Port); err != nil {
			s.log.Error("failed to server: " + err.Error())
		}
		return
	}()
	return nil
}

func (s *Server) OnStop(_ context.Context) error {
	s.log.Debug("stop server")
	return nil
}

var _ usecase.InterfaceUsecase = (*usecase.Usecase)(nil)
var _ InterfaceServer = (*Server)(nil)
