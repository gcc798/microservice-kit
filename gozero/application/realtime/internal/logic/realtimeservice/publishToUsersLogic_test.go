package realtimeservicelogic

import (
	"testing"

	"github.com/gcc798/microservice-kit/application/realtime/pb"
)

func TestValidatePublish(t *testing.T) {
	users, data, err := validatePublish(&pb.PublishToUsersReq{UserIds: []int64{2, 2, 3}, Type: "notice", DataJson: []byte(`{"ok":true}`)})
	if err != nil || len(users) != 2 || string(data) != `{"ok":true}` {
		t.Fatalf("validatePublish() = %v, %s, %v", users, data, err)
	}
	for _, request := range []*pb.PublishToUsersReq{nil, {Type: "x", UserIds: []int64{0}}, {Type: "", UserIds: []int64{1}}, {Type: "x", UserIds: []int64{1}, DataJson: []byte("{")}} {
		if _, _, err := validatePublish(request); err == nil {
			t.Fatalf("validatePublish(%v) unexpectedly succeeded", request)
		}
	}
}
