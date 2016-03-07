package service

// StringService provides operations on strings.
type Service interface {
	GetResponse() interface{}
	Handle(request interface{})
}

type StringService struct {
	response CountResponse
}

type CountRequest struct {
	S string `json:"s"`
}

type CountResponse struct {
	V int `json:"v"`
}

func (s *StringService) Count(value string) int {
	return len(value)
}

func (s *StringService) Handle(request interface{}) {
	req := request.(CountRequest)
	v := s.Count(req.S)
	s.response = CountResponse{v}
}

func (s *StringService) GetResponse() interface{} {
	return s.response
}
