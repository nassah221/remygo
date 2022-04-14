package swagger

import (
	"context"
	"net/http"
	"testing"
)

var (
	c                          *APIClient
	o                          AddOrganization
	hostDevice, clientDevice   AddUserDevice
	host, client               AddUser
	hostSession, clientSession AddSession
	hostuser, clientuser       User
	hostdevice, clientdevice   UserDevice
	hostsession, clientsession Session
	organization               Organization
	hostdevices                SessionDevice
)

func TestMain(_ *testing.T) {
	c = NewAPIClient(NewConfiguration())
	o = AddOrganization{Name: "hyper Data Computing", Description: "Software House in Karachi", Location: "karachi,Pakistan"}

	host = AddUser{Fname: "host", Lname: "test", Email: "host.test@hdc.com", Password: "host123", Mobile: "021111222333"}
	hostDevice = AddUserDevice{Type_: "Windows", Softwareversion: "10"}
	hostSession = AddSession{Identifier: "key123"}

	client = AddUser{Fname: "client", Lname: "test", Email: "client.test@hdc.com", Password: "client123", Mobile: "021111222333"}
	clientDevice = AddUserDevice{Type_: "Windows", Softwareversion: "11"}
	clientSession = AddSession{Identifier: "key321"}

}

func TestUser(t *testing.T) {
	t.Run("test create org endpoint", func(t *testing.T) {
		ctx := context.Background()
		org, res, err := c.OrganizationApi.Create(ctx, o)
		if err != nil {
			t.Fail()
		}
		if status := res.StatusCode; status != http.StatusCreated {
			t.Errorf("Status code is %v, expected %v", status, http.StatusCreated)
		}
		organization = org
		t.Logf("Organization: %v", org)
		res.Body.Close()
	})

	t.Run("test Create host user endpoint", func(t *testing.T) {
		ctx := context.Background()
		u, res, err := c.UserApi.Create(ctx, organization.Id, host)
		if err != nil {
			t.Fail()
		}

		if status := res.StatusCode; status != http.StatusCreated {
			t.Errorf("Status code is %v, expected %v", status, http.StatusCreated)
		}
		hostuser = u
		t.Logf("Host User: %v", u)
		res.Body.Close()
	})

	t.Run("test Create client user endpoint", func(t *testing.T) {
		ctx := context.Background()
		u, res, err := c.UserApi.Create(ctx, organization.Id, client)
		if err != nil {
			t.Fail()
		}

		if status := res.StatusCode; status != http.StatusCreated {
			t.Errorf("Status code is %v, expected %v", status, http.StatusCreated)
		}
		clientuser = u
		t.Logf("Client User: %v", u)
		res.Body.Close()
	})

	t.Run("test Create host device endpoint", func(t *testing.T) {
		ctx := context.Background()
		d, res, err := c.UserDeviceApi.Create(ctx, hostuser.Id, hostDevice)
		if err != nil {
			t.Fail()
		}

		if status := res.StatusCode; status != http.StatusCreated {
			t.Errorf("Status code is %v, expected %v", status, http.StatusCreated)
		}
		t.Logf("Host device: %v", d)
		res.Body.Close()
	})
	t.Run("test Create client device endpoint", func(t *testing.T) {
		ctx := context.Background()
		d, res, err := c.UserDeviceApi.Create(ctx, clientuser.Id, clientDevice)
		if err != nil {
			t.Fail()
		}

		if status := res.StatusCode; status != http.StatusCreated {
			t.Errorf("Status code is %v, expected %v", status, http.StatusCreated)
		}
		t.Logf("Client Device: %v", d)
		res.Body.Close()
	})
	// t.Run("test Create host Session endpoint", func(t *testing.T) {
	// 	ctx := context.Background()
	// 	s, res, err := c.SessionApi.Create(ctx, hostdevice.Id, hostSession)
	// 	if err != nil {
	// 		t.Fail()
	// 	}

	// 	if status := res.StatusCode; status != http.StatusCreated {
	// 		t.Errorf("Status code is %v, expected %v", status, http.StatusCreated)
	// 	}
	// 	hostsession = s
	// 	t.Logf("Host Session: %v", hostsession)
	// 	res.Body.Close()
	// })

	// t.Run("test Create client Session endpoint", func(t *testing.T) {
	// 	ctx := context.Background()
	// 	s, res, err := c.SessionApi.Create(ctx, clientdevice.Id, clientSession)
	// 	if err != nil {
	// 		t.Fail()
	// 	}

	// 	if status := res.StatusCode; status != http.StatusCreated {
	// 		t.Errorf("Status code is %v, expected %v", status, http.StatusCreated)
	// 	}
	// 	clientsession = s
	// 	t.Logf("Client Session: %v", clientsession)
	// 	res.Body.Close()
	// })

	// t.Run("test client join host Session endpoint", func(t *testing.T) {
	// 	ctx := context.Background()
	// 	hs, res, err := c.SessionDeviceApi.Create(ctx, hostsession.Identifier, clientdevice.Id)
	// 	if err != nil {
	// 		t.Fail()
	// 	}

	// 	if status := res.StatusCode; status != http.StatusCreated {
	// 		t.Errorf("Status code is %v, expected %v", status, http.StatusCreated)
	// 	}
	// 	hostdevices = hs
	// 	t.Logf("host devices: %v", hostdevices)
	// 	res.Body.Close()
	// })
}
