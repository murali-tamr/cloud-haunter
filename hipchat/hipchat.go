package hipchat

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"

	log "github.com/Sirupsen/logrus"
	ctx "github.com/hortonworks/cloud-haunter/context"
	"github.com/hortonworks/cloud-haunter/types"
	"github.com/tbruyelle/hipchat-go/hipchat"
)

var dispatcher = hipchatDispatcher{}

type hipchatDispatcher struct {
	client *hipchat.Client
	room   string
}

func init() {
	token := os.Getenv("HIPCHAT_TOKEN")
	server := os.Getenv("HIPCHAT_SERVER")
	room := os.Getenv("HIPCHAT_ROOM")
	if len(token) == 0 || len(server) == 0 || len(room) == 0 {
		log.Warn("[HIPCHAT] HIPCHAT_TOKEN or HIPCHAT_SERVER or HIPCHAT_ROOM environment variables are missing")
		return
	}
	if serverURL, err := url.Parse(server); err != nil {
		log.Errorf("[HIPCHAT] invalid url: %s, err: %s", server, err.Error())
	} else {
		log.Infof("[HIPCHAT] register hipchat to send notifications to server: %s and room: %s", serverURL.Host, room)
		dispatcher.init(token, room, serverURL)
		ctx.Dispatchers["HIPCHAT"] = dispatcher
	}
}

func (d *hipchatDispatcher) init(token, room string, serverURL *url.URL) {
	d.room = room
	d.client = hipchat.NewClient(token)
	d.client.BaseURL = serverURL
}

func (d hipchatDispatcher) GetName() string {
	return "HipChat"
}

func (d hipchatDispatcher) Send(op *types.OpType, items []types.CloudItem) error {
	message := d.generateMessage(op, items)
	log.Debugf("[HIPCHAT] Generated message is: %s", message)
	if ctx.DryRun {
		log.Info("[HIPCHAT] Skipping notification on dry run session")
	} else {
		return send(d.room, message, d.client.Room)
	}
	return nil
}

func (d *hipchatDispatcher) generateMessage(op *types.OpType, items []types.CloudItem) string {
	var buffer bytes.Buffer
	buffer.WriteString("/code\n")
	for _, item := range items {
		switch item.GetItem().(type) {
		case types.Instance:
			inst := item.GetItem().(types.Instance)
			buffer.WriteString(fmt.Sprintf("[%s] instance name: %s type: %s created: %s owner: %s region: %s\n", item.GetCloudType(), item.GetName(), inst.InstanceType, item.GetCreated(), item.GetOwner(), inst.Region))
		default:
			buffer.WriteString(fmt.Sprintf("[%s] %s: %s created: %s owner: %s\n", item.GetCloudType(), item.GetType(), item.GetName(), item.GetCreated(), item.GetOwner()))
		}
	}
	return buffer.String()
}

type notificationClient interface {
	Notification(string, *hipchat.NotificationRequest) (*http.Response, error)
}

func send(room, message string, client notificationClient) error {
	_, err := client.Notification(room, &hipchat.NotificationRequest{
		Message:       message,
		Color:         "green",
		MessageFormat: "text",
	})
	return err
}
