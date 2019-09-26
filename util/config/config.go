// Copyright (C) 2018 go-gamc authors
//
// This file is part of the go-gamc library.
//
// the go-gamc library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-gamc library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-gamc library.  If not, see <http://www.gnu.org/licenses/>.
//

package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"sync"
)

const (
	DefaultConfigPath = "conf/conf.yaml"
)

var _config *Config

type Config struct {
	data     []byte
	cache    map[string]interface{}
	filePath string
	mu       sync.Mutex
}

func Get(key string) interface{} {
	conf, err := InitConfig("")
	if err == nil {
		return conf.Get(key)
	}
	return nil
}

func GetInt(key string) int {
	conf, err := InitConfig("")
	if err == nil {
		return conf.GetInt(key)
	}
	return 0
}

func GetString(key string) string {
	conf, err := InitConfig("")
	if err == nil {
		return conf.GetString(key)
	}
	return ""
}

func (c *Config) GetString(key string) string {
	v := c.Get(key)
	if v != nil && reflect.TypeOf(v).Kind() == reflect.String {
		return v.(string)
	}
	return ""
}

func (c *Config) GetInt(key string) int {
	v := c.Get(key)
	if v != nil && reflect.TypeOf(v).Kind() == reflect.Int {
		return v.(int)
	}
	return 0
}

func (c *Config) Get(key string) interface{} {
	for keyName, _ := range c.cache {
		if keyName == key {
			return c.cache[key]
		}
	}

	keys := strings.Split(key, "/")
	if len(keys) == 0 {
		return nil
	}

	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(c.data, m)
	if err != nil {
		return nil
	}
	for i := 0; i < len(keys); i++ {
		if m[keys[i]] != nil {
			if i < len(keys)-1 {
				if reflect.TypeOf(m[keys[i]]).Kind() == reflect.Map {
					m = (m[keys[i]]).(map[interface{}]interface{})
				} else {
					return nil
				}
			} else {
				v := m[keys[len(keys)-1]]
				c.cache[key] = v
				return v
			}
		} else {
			return nil
		}
	}
	return nil
}

func (c *Config) GetObject(key string, destObj interface{}) interface{} {
	keys := strings.Split(key, "/")
	if len(keys) == 0 {
		return nil
	}

	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(c.data, m)
	if err != nil {
		return nil
	}
	for i := 0; i < len(keys); i++ {
		if m[keys[i]] != nil {
			if i < len(keys)-1 {
				if reflect.TypeOf(m[keys[i]]).Kind() == reflect.Map {
					m = (m[keys[i]]).(map[interface{}]interface{})
				} else {
					return nil
				}
			} else {
				out, err := yaml.Marshal(m[keys[i]])
				if err == nil {
					err = yaml.Unmarshal(out, destObj)
					if err != nil {
						log.Println(err)
					}
					return destObj
				}
			}
		} else {
			return nil
		}
	}
	return nil
}

func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := strings.Split(key, "/")
	if len(keys) == 0 {
		return
	}
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(c.data, m)
	if err != nil {
		return
	}
	tempMap := m
	for i := 0; i < len(keys); i++ {
		if i == len(keys)-1 {
			tempMap[keys[i]] = value
		} else {
			if tempMap[keys[i]] == nil {
				tempMap[keys[i]] = make(map[interface{}]interface{})
			}
			tempMap = (tempMap[keys[i]]).(map[interface{}]interface{})
		}
	}
	tempData, err := yaml.Marshal(m)
	if err == nil {
		c.data = tempData
	}
}

func InitConfig(fileName string) (*Config, error) {
	if _config != nil {
		return _config, nil
	}
	if fileName == "" {
		fileName = DefaultConfigPath
	}
	in, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic("Failed to read the config file:" + fileName + ". err:" + err.Error())
	}
	_config = new(Config)
	_config.filePath = fileName
	_config.data = in
	_config.cache = make(map[string]interface{})
	return _config, nil
}
