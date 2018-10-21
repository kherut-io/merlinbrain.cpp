package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/robfig/cron"
)

func redirectDashboard(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard/", 301)
}

func runCmd(cmd string) string {
	out, err := exec.Command("sh", "-c", cmd).Output()
	_ = err

	return string(out)
}

func main() {
	c := cron.New()

	counter := 0

	c.AddFunc("0 0 * * * *", func() {
		currentTime := time.Now()

		cmd := "echo " + currentTime.Format("2006.01.02-15:04:05") + "#" + strings.Split(runCmd("cat /sys/devices/virtual/thermal/thermal_zone0/temp"), "\n")[0] + " >> ./data/temperature.data"

		if counter > 255 {
			runCmd("tail -n 255 data/temperature.data > data/temperature.tmp && mv data/temperature.tmp data/temperature.data")
		}

		runCmd(cmd)

		counter += 1
	})

	c.Start()

	//DASHBOARD AT /dashboard
	http.Handle("/dashboard/", http.StripPrefix("/dashboard/", http.FileServer(http.Dir("template"))))

	//CONTROL AT /control
	http.HandleFunc("/control/", func(w http.ResponseWriter, r *http.Request) {
		command := strings.Join(strings.Split(r.URL.Path[1:], "/")[1:], "/")

		category := strings.Split(command, "/")[0]

		var arg []string

		if len(strings.Split(command, "/")) > 1 {
			arg = strings.Split(command, "/")[1:]
		}

		if category == "info" {
			switch arg[0] {
			case "temperature":
				if len(arg) > 1 {
					if arg[1] == "all" {
						w.Write([]byte(runCmd("cat ./data/temperature.data")))
					}
				} else {
					w.Write([]byte(runCmd("cat /sys/devices/virtual/thermal/thermal_zone0/temp")))
				}
			case "uptime":
				w.Write([]byte(strings.Split(runCmd("cat /proc/uptime"), ".")[0]))
			}
		} else if category == "led" {
			var cmd string

			switch arg[0] {
			case "on":
				cmd = "echo 1 > /sys/class/leds/red_led/brightness"

			case "off":
				cmd = "echo 0 > /sys/class/leds/red_led/brightness"
			}

			output := runCmd(cmd)

			if len(output) > 0 {
				w.Write([]byte("\n\nOutput: " + output))
				fmt.Println(output)
			}
		}
	})

	http.HandleFunc("/", redirectDashboard)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
