package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/micro/go-micro/v2/logger"
	"log"
	"net"
	"net/url"
	"test/monitor"
)

type notificationSubscriber struct {
	cluster uint64
}

// Start openstack notification Start
func (n *notification) Start(ctx context.Context) error {
	if err := n.Connect(); err != nil {
		return err
	}

	logger.Infof("Success to connect cluster notification.")

	go func() {
		defer func() {
			_ = n.Disconnect()
		}()

		ns := notificationSubscriber{
			cluster: 1,
		}
		s, err := n.Subscribe(ns.subscribeEvent)
		if err != nil {
			logger.Errorf("Could not register to cluster notification. Cause: %v", err)
			return
		}
		defer func() {
			_ = s.Unsubscribe()
		}()

		logger.Info("Start cluster notification.\n")
		<-ctx.Done()
	}()

	return nil
}

// Stop openstack notification stop
func (n *notification) Stop() {
	_ = n.Disconnect()
	logger.Infof("Close cluster notification.")
}

func (ns *notificationSubscriber) subscribeEvent(p Event) error {
	var e map[string]interface{}

	err := json.Unmarshal(p.Message().Body, &e)
	if err != nil {
		log.Printf("Could not subscribe event. cause: %v\n", err)
		return err
	}

	var m OsloMessage
	if err := json.Unmarshal([]byte(e["oslo.message"].(string)), &m); err != nil {
		log.Printf("Could not subscribe event. cause: %v\n", err)
		return err
	}

	//TODO message type 별 동기화진행
	switch m.EventType {
	case "identity.project.created":
		fallthrough
	case "identity.project.updated":
		fallthrough
	case "identity.project.deleted":
		log.Printf("project notification %s\n", m.Payload["id"].(string))
	case "compute.instance.create.end":
		fallthrough
	case "compute.instance.update":
		fallthrough
	case "compute.instance.delete.end":
		fallthrough
	case "compute.instance.suspend.end":
		log.Printf("instance notification %s\n", m.Payload["instance_id"].(string))
	case "volume.attach.end":

	case "volume.create.end":
		fallthrough
	case "volume.update.end":
		fallthrough
	case "volume.delete.end":
		log.Printf("volume notification %s\n", m.Payload["volume"].(string))
	case "snapshot.create.end":
		fallthrough
	case "snapshot.update.end":
		fallthrough
	case "snapshot.delete.end":
		log.Printf("snapshot notification %s\n", m.Payload["snapshot_id"].(string))
	case "volume_type.create":
		fallthrough
	case "volume_type.update":
		fallthrough
	case "volume_type.delete":
		log.Printf("storage notification %s\n", m.Payload["volume_types"].(map[string]interface{})["id"].(string))
	case "volume_type_project.access.add":
	case "volume_type_extra_specs.create":
	case "volume_type_extra_specs.delete":
	case "network.create.end":
		fallthrough
	case "network.update.end":
		fallthrough
	case "network.delete.end":
		log.Printf("network notification %s\n", m.Payload["network"].(map[string]interface{})["id"].(string))
	case "subnet.create.end":
		fallthrough
	case "subnet.update.end":
		log.Printf("subnet notification %s\n", m.Payload["subnet"].(map[string]interface{})["id"].(string))
	case "security_group.create.end":
		fallthrough
	case "security_group.update.end":
		fallthrough
	case "security_group.delete.end":
		log.Printf("sg notification %s\n", m.Payload["security_group"].(map[string]interface{})["id"].(string))
	case "security_group_rule.create.end":
		fallthrough
	case "security_group_rule.update.end":
		fallthrough
	case "security_group_rule.delete.end":
		log.Printf("sg rule notification %s\n", m.Payload["security_group_rule"].(map[string]interface{})["id"].(string))
	case "router.create.end":
		fallthrough
	case "router.update.end":
		fallthrough
	case "router.delete.end":
		log.Printf("router notification %s\n", m.Payload["router"].(map[string]interface{})["id"].(string))
	case "router.interface.create":
	case "floatingip.create.end":
		fallthrough
	case "floatingip.update.end":
		fallthrough
	case "floatingip.delete.end":
		log.Printf("floating ip notification %s\n", m.Payload["floatingip"].(map[string]interface{})["id"].(string))
	}
	if err != nil {
		log.Printf("Failed to sync cluster from event notification. cause: %v\n", err)
		return nil
	}

	log.Printf("success to sync cluster from event notification.\n")
	return nil
}

// New 함수는 새로운 monitor interface 를 생성한다.
func New(serverURL string) monitor.Monitor {
	//TODO auth 의 경우 임시로 ID:PASSWORD(ex.guest:guest)를 쓰지만
	//	   사용자 입력에 의한 Cluster 의 MetaData 로 저장될 필요가 있음.
	//     마찬가지로 임시로 client 의 api server url 과 고정된 port(ex.192.168.1.1:5672) 를 쓰지만
	//     사용자 입력에 의한 Cluster 의 MetaData 로 저장될 필요가 있음.
	auth := "guest:guest"
	defaultPort := "5672"

	u, _ := url.Parse(serverURL)
	ip, _, _ := net.SplitHostPort(u.Host)

	return &notification{
		auth:    auth,
		address: fmt.Sprintf("%s:%s", ip, defaultPort),
	}
}