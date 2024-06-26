package p2p

import (
	"context"
	"errors"
	"reflect"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/toknowwhy/theunit-oracle/internal/p2p/sets"
	"github.com/toknowwhy/theunit-oracle/pkg/transport"
)

var ErrNilMessage = errors.New("message is nil")

type Subscription struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	topic          *pubsub.Topic
	sub            *pubsub.Subscription
	teh            *pubsub.TopicEventHandler
	validatorSet   *sets.ValidatorSet
	eventHandler   sets.PubSubEventHandler
	messageHandler sets.MessageHandler

	// msgCh is used to send a notification about a new message, it's
	// returned by the Transport.Messages function.
	msgCh chan transport.ReceivedMessage
}

func newSubscription(node *Node, topic string, typ transport.Message) (*Subscription, error) {
	var err error
	ctx, ctxCancel := context.WithCancel(node.ctx)
	s := &Subscription{
		ctx:            ctx,
		ctxCancel:      ctxCancel,
		validatorSet:   node.validatorSet,
		eventHandler:   node.pubSubEventHandlerSet,
		messageHandler: node.messageHandlerSet,
		msgCh:          make(chan transport.ReceivedMessage),
	}
	err = node.pubSub.RegisterTopicValidator(topic, s.validator(topic, typ))
	if err != nil {
		return nil, err
	}
	s.topic, err = node.PubSub().Join(topic)
	if err != nil {
		return nil, err
	}
	s.sub, err = s.topic.Subscribe()
	if err != nil {
		return nil, err
	}
	s.teh, err = s.topic.EventHandler()
	if err != nil {
		return nil, err
	}
	s.messageLoop()
	s.eventLoop()
	return s, err
}

func (s *Subscription) Publish(message transport.Message) error {
	b, err := message.Marshall()
	if err != nil {
		return err
	}
	if b == nil {
		return ErrNilMessage
	}
	s.messageHandler.Published(s.topic.String(), b, message)
	return s.topic.Publish(s.ctx, b)
}

func (s *Subscription) Next() chan transport.ReceivedMessage {
	return s.msgCh
}

func (s *Subscription) validator(topic string, typ transport.Message) pubsub.ValidatorEx {
	// Validator actually have two roles in the libp2p: it unmarshalls messages
	// and then validates them. Unmarshalled message is stored in the
	// ValidatorData field which was created for this purpose:
	// https://github.com/libp2p/go-libp2p-pubsub/pull/231
	r := reflect.TypeOf(typ).Elem()
	return func(ctx context.Context, id peer.ID, psMsg *pubsub.Message) pubsub.ValidationResult {
		msg := reflect.New(r).Interface().(transport.Message)
		err := msg.Unmarshall(psMsg.Data)
		if err != nil {
			s.messageHandler.Broken(topic, psMsg, err)
			return pubsub.ValidationReject
		}
		psMsg.ValidatorData = msg
		vr := s.validatorSet.Validator(topic)(ctx, id, psMsg)
		s.messageHandler.Received(topic, psMsg, vr)
		return vr
	}
}

func (s *Subscription) messageLoop() {
	go func() {
		for {
			var msg transport.Message
			psMsg, err := s.sub.Next(s.ctx)

			if psMsg != nil && err == nil {
				msg = psMsg.ValidatorData.(transport.Message)
			}
			select {
			case <-s.ctx.Done():
				close(s.msgCh)
				return
			case s.msgCh <- transport.ReceivedMessage{
				Message: msg,
				Data:    psMsg,
				Error:   err,
			}:
			}
		}
	}()
}

func (s *Subscription) eventLoop() {
	go func() {
		for {
			pe, err := s.teh.NextPeerEvent(s.ctx)
			if err != nil {
				// The only time when an error may be returned here is
				// when the subscription is canceled.
				return
			}
			s.eventHandler.Handle(s.topic.String(), pe)
		}
	}()
}

func (s *Subscription) close() error {
	s.ctxCancel()
	s.teh.Cancel()
	s.sub.Cancel()
	return s.topic.Close()
}
