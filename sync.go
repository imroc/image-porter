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

func Sync(config *Config) error {
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
			klog.Errorf("failed to list dst tags for image %s: %v", image.From, err)
			continue
		}
		tags := tagsToSync(srcTags, dstTags)
		if len(tags) == 0 {
			klog.Infof("skip: all tags are already been synced for image %s", image.From)
			continue
		}
		klog.Infof("sync tags for image %s: %v", image.From, tags)
		err = syncImageTags(image.From, image.To, tags)
		if len(tags) == 0 {
			klog.Errorf("failed to sync image %s: %v", image.From, err)
			continue
		}
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
		if !dstMap[tag] {
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

func syncImageTags(imageFrom, imageTo string, tags []string) error {
	for _, tag := range tags {
		args := []string{"copy", "--all", fmt.Sprintf("docker://%s:%s", imageFrom, tag), fmt.Sprintf("docker://%s:%s", imageTo, tag)}
		klog.Infof("execute command: skopeo %s", strings.Join(args, " "))
		cmd := exec.Command("skopeo", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
