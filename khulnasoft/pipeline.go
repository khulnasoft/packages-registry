package khulnasoft

import (
	"fmt"
	"strings"
)

// PipelineJob represents a KhulnaSoft pipeline job.
type PipelineJob interface {
	Label() string // Returns the label of this job.
}

// Pipeline represents a KhulnaSoft pipeline configuration, which can be summarized by a
// collection of jobs.
type Pipeline struct {
	Stages []string
	Jobs   map[string]PipelineJob `yaml:",inline"`
}

// AddHiddenJob adds a hidden job definition to the pipeline. That job is defined by
// a stage, an image and a set of scripts. See https://docs.khulnasoft.com/ee/ci/jobs/#hide-jobs
func (p *Pipeline) AddHiddenJob(stage, image string, scripts []string) error {
	label := p.hiddenJobLabel(stage)
	hiddenJob, err := newHiddenJob(label, stage, image, scripts)
	if err != nil {
		return err
	}

	p.addPipelineElement(hiddenJob)

	return nil
}

type standardJob struct {
	stage                  string
	image                  string
	packageName            string
	packageVersion         string
	additionalEnvVariables map[string]string
}

// AddJob adds a standard job. Since that job is specialized for a package import, it is defined by
// a stage, an image, a package name and a version. Callers can pass additional environment variables
// that will be set in the job
func (p *Pipeline) AddJob(options ...func(*standardJob)) {
	job := &standardJob{}

	for _, o := range options {
		o(job)
	}

	variables := p.packageVariables(job.packageName, job.packageVersion, job.additionalEnvVariables)
	label := fmt.Sprintf("%s:%s:%s", job.stage, job.packageName, job.packageVersion)
	element := newJob(label, p.hiddenJobLabel(job.stage), variables)

	p.addPipelineElement(element)
}

func (p *Pipeline) withStage(stage string) func(*standardJob) {
	return func(s *standardJob) {
		s.stage = stage
	}
}

func (p *Pipeline) withImage(image string) func(*standardJob) {
	return func(s *standardJob) {
		s.image = image
	}
}

func (p *Pipeline) withPackageNameAndVersion(name, version string) func(*standardJob) {
	return func(s *standardJob) {
		s.packageName = name
		s.packageVersion = version
	}
}

func (p *Pipeline) withAdditionalEnvVariables(vars map[string]string) func(*standardJob) {
	return func(s *standardJob) {
		s.additionalEnvVariables = vars
	}
}

func (p *Pipeline) addPipelineElement(pj PipelineJob) {
	p.Jobs[pj.Label()] = pj
}

func (p *Pipeline) hiddenJobLabel(stage string) string {
	return fmt.Sprintf(".%s:scripts", stage)
}

func (p *Pipeline) packageVariables(pkgname, pkgversion string, additionaEnvlVariables map[string]string) map[string]string {
	vars := map[string]string{
		"PACKAGE_NAME":    pkgname,
		"PACKAGE_VERSION": pkgversion,
	}
	for key, value := range additionaEnvlVariables {
		vars[key] = value
	}
	return vars
}

func newPipeline(stageSize, jobsSize int) *Pipeline {
	return &Pipeline{
		Stages: make([]string, 0, stageSize),
		Jobs:   make(map[string]PipelineJob, jobsSize),
	}
}

// Job represents a job for a KhulnaSoft pipeline. Since the scripts are centralized in an
// job, this structure only needs to define the environment variables that those scripts need.
type Job struct {
	label     string
	Extends   string
	Variables map[string]string
}

// Label returns the job's label.
func (j Job) Label() string {
	return j.label
}

func newJob(label, extends string, variables map[string]string) Job {
	return Job{
		label:     label,
		Extends:   extends,
		Variables: variables,
	}
}

// HiddenJob represents a hidden job for a KhulnaSoft pipeline. See https://docs.khulnasoft.com/ee/ci/jobs/#hide-jobs.
// Hidden jobs will mainly host the scripts lines that need to be executed to import a package.
type HiddenJob struct {
	label   string
	Image   string
	Stage   string
	Needs   []string
	Scripts []string `yaml:"script"`
}

// Label returns the hidden job's label.
func (j HiddenJob) Label() string {
	return j.label
}

func newHiddenJob(label, stage, image string, scripts []string) (HiddenJob, error) {
	if !strings.HasPrefix(label, ".") {
		return HiddenJob{}, fmt.Errorf("error with label %q: hidden jobs labels must start by a dot(.)", label)
	}

	return HiddenJob{
		label:   label,
		Stage:   stage,
		Image:   image,
		Needs:   []string{},
		Scripts: scripts,
	}, nil
}
