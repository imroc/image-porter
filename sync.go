package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

type Tags struct {
	Tags []string `json:"Tags"`
}

func Sync(config *Config, retryAttempt int) error {
	for _, image := range config.Images {
		srcTags, err := listTags(image.From)
		if err != nil {
			klog.Errorf("failed to list src tags for image %s: %v", image.From, err)
			continue
		}
		klog.V(2).Infof("full tag list for image %s: %v", image.From, srcTags)
		srcTags = image.TagFilter.FilterTags(srcTags)
		if len(srcTags) == 0 {
			klog.Infof("no tags matched for image %s after filter", image.From)
			continue
		}
		klog.V(2).Infof("filterd tag list for image %s (%s): %v", image.From, image.TagFilter.String(), srcTags)
		dstTags, err := listTags(image.To)
		if err != nil {
			klog.Infof("dst image have no tags: %s (%v)", image.To, err)
		}
		tags := tagsToSync(srcTags, dstTags)
		if len(tags) == 0 {
			klog.Infof("skip: all tags are already been synced for image %s", image.From)
			continue
		}
		klog.Infof("sync tags for image %s: %v", image.From, tags)
		syncImageTags(image.From, image.To, tags, retryAttempt)
	}
	return nil
}

func listTags(image string) ([]string, error) {
	output, err := RunCommand("skopeo", "list-tags", fmt.Sprintf(`docker://%s`, image))
	if err != nil {
		return nil, err
	}
	var tags Tags
	err = json.Unmarshal(output, &tags)
	if err != nil {
		return nil, err
	}
	return tags.Tags, nil
}

func tagsToSync(srcTags, dstTags []string) []string {
	result := []string{}
	if len(dstTags) == 0 {
		return srcTags
	}
	dstMap := sliceToMap(dstTags)
	for _, tag := range srcTags {
		if tag == "latest" || !dstMap[tag] {
			result = append(result, tag)
		}
	}
	return result
}

func sliceToMap[T comparable](slice []T) map[T]bool {
	m := make(map[T]bool)
	for _, v := range slice {
		m[v] = true
	}
	return m
}

func syncImageTags(imageFrom, imageTo string, tags []string, retryAttempt int) {
	for _, tag := range tags {
		args := []string{"copy", "--all", fmt.Sprintf("docker://%s:%s", imageFrom, tag), fmt.Sprintf("docker://%s:%s", imageTo, tag)}
		klog.Infof("execute command: skopeo %s", strings.Join(args, " "))
		cmd := exec.Command("skopeo", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		var err error
		for i := 0; i < retryAttempt; i++ {
			err = cmd.Run()
			if err == nil {
				continue
			}
		}
	}
}
