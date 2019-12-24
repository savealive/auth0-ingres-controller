package auth0utils

import (
	"gopkg.in/auth0.v1/management"
	"reflect"
	"testing"
)

func TestAddItem(t *testing.T) {
	type args struct {
		s    *[]interface{}
		elem []string
	}
	tests := []struct {
		name      string
		args      args
		wantAdded int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAdded := AddItem(tt.args.s, tt.args.elem...); gotAdded != tt.wantAdded {
				t.Errorf("AddItem() = %v, want %v", gotAdded, tt.wantAdded)
			}
		})
	}
}

func TestClientList_AppID1(t *testing.T) {
	m, err := management.New(
		"creditplace.eu.auth0.com",
		"Anex52oIVnMlSmdOXU2yPXoYW70XqLcT",
		"hqVrLg5KvVpeDEM9h2PTUtT1ErKWpyx7q09rTp2-kMVDvpZTrn7ax0CkFgurAC3Q",
	)

	if err != nil {
		t.Fatal(err)
	}

	apps, err := m.Client.List()

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		l       ClientList
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "GOOD",
			l:       apps,
			args:    args{name: "Delete ME"},
			want:    "r4qFLEDP0flYjGPH21FsVgR14w4SSJY1",
			wantErr: false,
		},
		{
			name:    "BAD",
			l:       apps,
			args:    args{name: "Fake"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.AppID(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AppID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteItem(t *testing.T) {
	type args struct {
		s    *[]interface{}
		elem []string
	}
	tests := []struct {
		name        string
		args        args
		wantDeleted int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDeleted := DeleteItem(tt.args.s, tt.args.elem...); gotDeleted != tt.wantDeleted {
				t.Errorf("DeleteItem() = %v, want %v", gotDeleted, tt.wantDeleted)
			}
		})
	}
}

func TestSortUniq(t *testing.T) {
	type args struct {
		s []interface{}
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SortUniq(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortUniq() = %v, want %v", got, tt.want)
			}
		})
	}
}
