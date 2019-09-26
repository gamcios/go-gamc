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

package pprof

import (
	"gamc.pro/gamcio/go-gamc/util/config"
	"gamc.pro/gamcio/go-gamc/util/logging"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"runtime/pprof"
)

type PprofConfig struct {
	HttpListen string `yaml:"http_listen"`
	Cpuprofile string `yaml:"cpuprofile"`
	Memprofile string `yaml:"memprofile"`
}

func GetPprofConfig(conf *config.Config) *PprofConfig {
	pprofcfg := new(PprofConfig)
	conf.GetObject("monitor/pprof", pprofcfg)
	return pprofcfg
}

type Pprof struct {
	Config *PprofConfig
}

// StartProfiling try start pprof
func (p *Pprof) StartProfiling() {

	if p.Config == nil {
		logging.CLog().Error("Failed to find monitor.pprof config in config file")
		return
	}

	cpuProfile := p.Config.Cpuprofile
	if len(cpuProfile) > 0 {
		f, err := os.Create(cpuProfile)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err": err,
			}).Fatalf("Failed to create CPU profile %s", cpuProfile)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err": err,
			}).Fatalf("Failed to start CPU profile")
		}
	}
}

// StopProfiling try stop pprof
func (p *Pprof) StopProfiling() {
	if p.Config == nil {
		return
	}

	memProfile := p.Config.Memprofile
	if len(memProfile) > 0 {
		f, err := os.Create(memProfile)
		if err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err": err,
			}).Errorf("Failed to create memory profile %s", memProfile)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			logging.CLog().WithFields(logrus.Fields{
				"err": err,
			}).Errorf("Failed to write memory profile")
		}
		f.Close()
	}

	cpuProfile := p.Config.Cpuprofile
	if len(cpuProfile) > 0 {
		pprof.StopCPUProfile()
	}
}
