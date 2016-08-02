package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/screwdriver-cd/launcher/screwdriver"
)

type NewAPI func(buildID string, token string) (screwdriver.API, error)
type FakeAPI screwdriver.API
type FakeBuild screwdriver.Build
type FakeJob screwdriver.Job
type FakePipeline screwdriver.Pipeline

type MockAPI struct {
	buildFromID    func(string) (screwdriver.Build, error)
	jobFromID      func(string) (screwdriver.Job, error)
	pipelineFromID func(string) (screwdriver.Pipeline, error)
}

func (f MockAPI) BuildFromID(buildID string) (screwdriver.Build, error) {
	if f.buildFromID != nil {
		return f.buildFromID(buildID)
	}
	return screwdriver.Build(FakeBuild{}), nil
}

func (f MockAPI) JobFromID(jobID string) (screwdriver.Job, error) {
	if f.jobFromID != nil {
		return f.jobFromID(jobID)
	}
	return screwdriver.Job(FakeJob{}), nil
}

func (f MockAPI) PipelineFromID(pipelineID string) (screwdriver.Pipeline, error) {
	if f.pipelineFromID != nil {
		return f.pipelineFromID(pipelineID)
	}
	return screwdriver.Pipeline(FakePipeline{}), nil
}

func TestMain(m *testing.M) {
	mkdirAll = func(path string, perm os.FileMode) (err error) { return nil }
	stat = func(path string) (info os.FileInfo, err error) { return nil, os.ErrExist }
}

func TestBuildFromId(t *testing.T) {
	testID := "TESTID"
	api := MockAPI{
		buildFromID: func(buildID string) (screwdriver.Build, error) {
			if buildID != testID {
				t.Errorf("buildID == %v, want %v", buildID, testID)
			}
			return screwdriver.Build(FakeBuild{}), nil
		},
	}

	launch(screwdriver.API(api), testID)
}

func TestBuildFromIdError(t *testing.T) {
	api := MockAPI{
		buildFromID: func(buildID string) (screwdriver.Build, error) {
			err := fmt.Errorf("testing error returns")
			return screwdriver.Build(FakeBuild{}), err
		},
	}

	err := launch(screwdriver.API(api), "shoulderror")
	if err == nil {
		t.Errorf("err should not be nil")
	}

	expected := `fetching build ID "shoulderror"`
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("err == %q, want %q", err, expected)
	}
}

func TestJobFromID(t *testing.T) {
	testBuildID := "BUILDID"
	testJobID := "JOBID"
	api := MockAPI{
		buildFromID: func(buildID string) (screwdriver.Build, error) {
			return screwdriver.Build(FakeBuild{ID: testBuildID, JobID: testJobID}), nil
		},
		jobFromID: func(jobID string) (screwdriver.Job, error) {
			if jobID != testJobID {
				t.Errorf("jobID == %v, want %v", jobID, testJobID)
			}
			return screwdriver.Job(FakeJob{}), nil
		},
	}

	launch(screwdriver.API(api), testBuildID)
}

func TestJobFromIdError(t *testing.T) {
	testBuildID := "BUILDID"
	testJobID := "JOBID"
	api := MockAPI{
		buildFromID: func(buildID string) (screwdriver.Build, error) {
			return screwdriver.Build(FakeBuild{ID: testBuildID, JobID: testJobID}), nil
		},
		jobFromID: func(jobID string) (screwdriver.Job, error) {
			err := fmt.Errorf("testing error returns")
			return screwdriver.Job(FakeJob{}), err
		},
	}

	err := launch(screwdriver.API(api), testBuildID)
	if err == nil {
		t.Errorf("err should not be nil")
	}

	expected := fmt.Sprintf(`fetching Job ID %q`, testJobID)
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("err == %q, want %q", err, expected)
	}
}

func TestPipelineFromID(t *testing.T) {
	testBuildID := "BUILDID"
	testJobID := "JOBID"
	testPipelineID := "PIPELINEID"
	api := MockAPI{
		buildFromID: func(buildID string) (screwdriver.Build, error) {
			return screwdriver.Build(FakeBuild{ID: testBuildID, JobID: testJobID}), nil
		},
		jobFromID: func(buildID string) (screwdriver.Job, error) {
			return screwdriver.Job(FakeJob{ID: testJobID, PipelineID: testPipelineID}), nil
		},
		pipelineFromID: func(pipelineID string) (screwdriver.Pipeline, error) {
			if pipelineID != testPipelineID {
				t.Errorf("pipelineID == %v, want %v", pipelineID, testPipelineID)
			}
			return screwdriver.Pipeline(FakePipeline{}), nil
		},
	}

	launch(screwdriver.API(api), testBuildID)
}

func TestPipelineFromIdError(t *testing.T) {
	testBuildID := "BUILDID"
	testJobID := "JOBID"
	testPipelineID := "PIPELINEID"
	api := MockAPI{
		buildFromID: func(buildID string) (screwdriver.Build, error) {
			return screwdriver.Build(FakeBuild{ID: testBuildID, JobID: testJobID}), nil
		},
		jobFromID: func(buildID string) (screwdriver.Job, error) {
			return screwdriver.Job(FakeJob{ID: testJobID, PipelineID: testPipelineID}), nil
		},
		pipelineFromID: func(pipelineID string) (screwdriver.Pipeline, error) {
			err := fmt.Errorf("testing error returns")
			return screwdriver.Pipeline(FakePipeline{}), err
		},
	}

	err := launch(screwdriver.API(api), testBuildID)
	if err == nil {
		t.Errorf("err should not be nil")
	}

	expected := fmt.Sprintf(`fetching Pipeline ID %q`, testPipelineID)
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("err == %q, want %q", err, expected)
	}
}

func TestParseScmURL(t *testing.T) {
	wantHost := "git@github.com"
	wantOrg := "screwdriver-cd"
	wantRepo := "launcher.git"
	wantBranch := "master"

	scmURL := "git@github.com:screwdriver-cd/launcher.git#master"
	parsedURL, err := parseScmURL(scmURL)
	host, org, repo, branch := parsedURL.Host, parsedURL.Org, parsedURL.Repo, parsedURL.Branch
	if err != nil {
		t.Errorf("Unexpected error parsing SCM URL %q: %v", scmURL, err)
	}

	if host != wantHost {
		t.Errorf("host = %q, want %q", host, wantHost)
	}

	if org != wantOrg {
		t.Errorf("org = %q, want %q", org, wantOrg)
	}

	if repo != wantRepo {
		t.Errorf("repo = %q, want %q", repo, wantRepo)
	}

	if branch != wantBranch {
		t.Errorf("branch = %q, want %q", branch, wantBranch)
	}

	if parsedURL.String() != scmURL {
		t.Errorf("parsedURL.String() == %q, want %q", parsedURL.String(), scmURL)
	}
}

func TestCreateWorkspace(t *testing.T) {
	testOrg := "screwdriver-cd"
	testRepo := "launcher.git"
	wantWorkspace := "/opt/screwdriver/workspace/src/screwdriver-cd/launcher.git"

	workspace, err := createWorkspace(testOrg, testRepo)
	if err != nil {
		t.Errorf("Unexpected error creating workspace: %v", err)
	}

	if workspace != wantWorkspace {
		t.Errorf("workspace = %q, want %q", workspace, wantWorkspace)
	}
}
