package app

import (
	"encoding/json"
	"fmt"
	"github.com/Kami-no/ward/src/config"
	"log"
	"net/http"
)

type MRController struct {
	Config *config.Config
}

func NewMRController(config *config.Config) *MRController {
	return &MRController{
		Config: config,
	}
}

func (c *MRController) Handler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "ok")
}

func (c *MRController) HandleMR(w http.ResponseWriter, _ *http.Request) {
	mrs, err := checkPrjRequests(c.Config, c.Config.Projects, "any")
	if err != nil {
		log.Println(err)
	}

	out, _ := json.Marshal(mrs)
	output := fmt.Sprintf("%v", string(out))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, output)
}

func (c *MRController) HandleMROpened(w http.ResponseWriter, _ *http.Request) {

	mrs, err := checkPrjRequests(c.Config, c.Config.Projects, "opened")
	if err != nil {
		log.Println(err)
	}

	actions := evalOpenedRequests(mrs)
	output := fmt.Sprintf("%v", actions)
	fmt.Fprint(w, output)
}

func (c *MRController) HandleMRMerged(w http.ResponseWriter, _ *http.Request) {
	mrs, err := checkPrjRequests(c.Config, c.Config.Projects, "merged")
	if err != nil {
		log.Println(err)
	}

	data := evalMergedRequests(mrs)

	out, _ := json.Marshal(data)
	output := fmt.Sprintf("%v", string(out))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, output)
}

func (c *MRController) HandleMRApply(w http.ResponseWriter, _ *http.Request) {
	data := DetectMR(c.Config)

	out, _ := json.Marshal(data)
	output := fmt.Sprintf("%v", string(out))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, output)
}

func (c *MRController) HandleDead(w http.ResponseWriter, _ *http.Request) {
	data := detectDead(c.Config)
	undead, _ := json.Marshal(data)

	output := fmt.Sprintf("%v", string(undead))

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, output)
}

func (c *MRController) HandleDeadLetter(w http.ResponseWriter, _ *http.Request) {
	undead := detectDead(c.Config)
	for _, v := range undead.Authors {
		v.Projects = undead.Projects
		template, err := deadAuthorTemplate(v)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		fmt.Fprint(w, template)
		break
	}
}
