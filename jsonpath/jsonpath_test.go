package jsonpath_test

import (
	"reflect"
	"testing"

	"github.com/shellyln/go-small-jsonpath/jsonpath"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		path    string
		want    interface{}
		wantErr bool
	}{{
		name:    "1",
		src:     `{"a":1}`,
		path:    `$.a`,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "2",
		src:     `{"a":1}`,
		path:    `$["a"]`,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "3",
		src:     `{"a":1}`,
		path:    `$['a']`,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "4",
		src:     "[5]",
		path:    `$[0]`,
		want:    float64(5),
		wantErr: false,
	}, {
		name:    "5",
		src:     `"a"`,
		path:    `$`,
		want:    "a",
		wantErr: false,
	}, {
		name:    "6",
		src:     "7",
		path:    `$`,
		want:    float64(7),
		wantErr: false,
	}, {
		name:    "7",
		src:     "-13",
		path:    `$`,
		want:    float64(-13),
		wantErr: false,
	}, {
		name:    "8",
		src:     "null",
		path:    `$`,
		want:    nil,
		wantErr: false,
	}, {
		name:    "9",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$.test[1].abc`,
		want:    float64(10),
		wantErr: false,
	}, {
		name:    "10",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$.test[1]["abc"]`,
		want:    float64(10),
		wantErr: false,
	}, {
		name:    "11",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$.test.(first).abc`,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "12",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$.test.(last)["abc"]`,
		want:    float64(10),
		wantErr: false,
	}, {
		name:    "13",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$.test.(length)`,
		want:    int(2),
		wantErr: false,
	}, {
		name:    "14",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$ . test [ 1 ] . abc `,
		want:    float64(10),
		wantErr: false,
	}, {
		name:    "15",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$ [ 'test' ] . (first) [ "abc" ] `,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "16",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$  [  'test'  ]  .  (first)  .  abc  `,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "17",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$ [ 'test' ] . (first) [ "\x61\x62\x63" ] `,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "18",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$ [ 'test' ] . (first) [ "\u0061\u0062\u0063" ] `,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "19",
		src:     `{"test":[{"abc":1},{"abc":10}]}`,
		path:    `$ [ 'test' ] . (first) [ "\u{0061}\u{00062}\u{000063}" ] `,
		want:    float64(1),
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json, err := jsonpath.ReadString(tt.src)
			if err != nil {
				t.Errorf("%v: ReadString: error = %v", tt.name, err)
				return
			}

			path, err := jsonpath.Compile(tt.path)
			if err != nil {
				t.Errorf("%v: Compile: error = %v", tt.name, err)
				return
			}

			v, err := path.Query(json)
			if err != nil {
				t.Errorf("%v: Query: error = %v", tt.name, err)
				return
			}

			if !reflect.DeepEqual(v, tt.want) {
				t.Errorf("%v: v = %v, want = %v", tt.name, v, tt.want)
				return
			}
		})
	}
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		path    string
		want    interface{}
		wantErr bool
	}{{
		name:    "1",
		src:     `{"a":1}`,
		path:    `$.a`,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "2",
		src:     `{"a":"x"}`,
		path:    `$.a`,
		want:    "x",
		wantErr: false,
	}, {
		name:    "3",
		src:     `{"a":[1]}`,
		path:    `$.a`,
		want:    []interface{}{float64(1)},
		wantErr: false,
	}, {
		name:    "4",
		src:     `{"a":{"b":1}}`,
		path:    `$.a`,
		want:    map[string]interface{}{"b": float64(1)},
		wantErr: false,
	}, {
		name:    "5",
		src:     `{"a":{"b":1}}`,
		path:    `$.c`,
		want:    nil,
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json, err := jsonpath.ReadString(tt.src)
			if err != nil {
				t.Errorf("%v: ReadString: error = %v", tt.name, err)
				return
			}

			path, err := jsonpath.Compile(tt.path)
			if err != nil {
				t.Errorf("%v: Compile: error = %v", tt.name, err)
				return
			}

			v, err := path.Query(json)
			if tt.wantErr {
				if err == nil {
					t.Errorf("%v: Query: want error: v = %v", tt.name, v)
					return
				}
			} else {
				if err != nil {
					t.Errorf("%v: Query: error = %v", tt.name, err)
					return
				}
			}

			if !reflect.DeepEqual(v, tt.want) {
				t.Errorf("%v: v = %v, want = %v", tt.name, v, tt.want)
				return
			}
		})
	}
}

func TestQueryAsNumberOrZero(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		path    string
		want    interface{}
		wantErr bool
	}{{
		name:    "1",
		src:     `{"a":1}`,
		path:    `$.a`,
		want:    float64(1),
		wantErr: false,
	}, {
		name:    "2",
		src:     `{"a":"x"}`,
		path:    `$.a`,
		want:    float64(0),
		wantErr: false,
	}, {
		name:    "3",
		src:     `{"a":[]}`,
		path:    `$.a`,
		want:    float64(0),
		wantErr: false,
	}, {
		name:    "4",
		src:     `{"a":{}}`,
		path:    `$.a`,
		want:    float64(0),
		wantErr: false,
	}, {
		name:    "5",
		src:     `{"a":{"b":1}}`,
		path:    `$.c`,
		want:    float64(0),
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json, err := jsonpath.ReadString(tt.src)
			if err != nil {
				t.Errorf("%v: ReadString: error = %v", tt.name, err)
				return
			}

			path, err := jsonpath.Compile(tt.path)
			if err != nil {
				t.Errorf("%v: Compile: error = %v", tt.name, err)
				return
			}

			v := path.QueryAsNumberOrZero(json)

			if !reflect.DeepEqual(v, tt.want) {
				t.Errorf("%v: v = %v, want = %v", tt.name, v, tt.want)
				return
			}
		})
	}
}

func TestQueryAsStringOrZero(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		path    string
		want    interface{}
		wantErr bool
	}{{
		name:    "1",
		src:     `{"a":1}`,
		path:    `$.a`,
		want:    "",
		wantErr: false,
	}, {
		name:    "2",
		src:     `{"a":"x"}`,
		path:    `$.a`,
		want:    "x",
		wantErr: false,
	}, {
		name:    "3",
		src:     `{"a":[]}`,
		path:    `$.a`,
		want:    "",
		wantErr: false,
	}, {
		name:    "4",
		src:     `{"a":{}}`,
		path:    `$.a`,
		want:    "",
		wantErr: false,
	}, {
		name:    "5",
		src:     `{"a":{"b":1}}`,
		path:    `$.c`,
		want:    "",
		wantErr: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json, err := jsonpath.ReadString(tt.src)
			if err != nil {
				t.Errorf("%v: ReadString: error = %v", tt.name, err)
				return
			}

			path, err := jsonpath.Compile(tt.path)
			if err != nil {
				t.Errorf("%v: Compile: error = %v", tt.name, err)
				return
			}

			v := path.QueryAsStringOrZero(json)

			if !reflect.DeepEqual(v, tt.want) {
				t.Errorf("%v: v = %v, want = %v", tt.name, v, tt.want)
				return
			}
		})
	}
}
