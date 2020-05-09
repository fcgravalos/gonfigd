package api

import (
	context "context"

	"github.com/fcgravalos/gonfigd/kv"
	"github.com/fcgravalos/gonfigd/pubsub"
	"github.com/rs/zerolog"
)

type server struct {
	kv.KV
	pubsub.PubSub
	zerolog.Logger
}

func (s *server) GetConfig(ctx context.Context, req *GetConfigRequest) (*GetConfigResponse, error) {
	cfg, err := s.Get(req.ConfigPath)
	if err != nil {
		s.Error().Msgf("error while trying to read %s: %v", req.ConfigPath, err)
		return nil, err
	}
	return &GetConfigResponse{Config: cfg.Text()}, nil
}

func (s *server) WatchConfig(req *WatchConfigRequest, stream Gonfig_WatchConfigServer) error {
	if !s.TopicExists(req.ConfigPath) {
		if err := s.CreateTopic(req.ConfigPath); err != nil {
			s.Error().Msgf("cannot subscribe to changes of %s: %v", req.ConfigPath, err)
			return err
		}
	}
	sID, sCh := s.Subscribe(req.ConfigPath)
	defer s.UnSubscribe(req.ConfigPath, sID)

	ctx := stream.Context()
	for {
		select {
		case ev := <-sCh:
			resp := &WatchConfigResponse{
				SubscriptionID: sID,
				Event:          ev.String(),
			}
			if err := stream.Send(resp); err != nil {
				s.Error().Msgf("failed to send response %v trhough stream: %v", resp, err)
				return err
			}
			s.Info().Msgf("event %s sent to subscription ID %s", resp.Event, resp.SubscriptionID)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewServer(kv kv.KV, ps pubsub.PubSub, logger zerolog.Logger) *server {
	return &server{kv, ps, logger}
}
