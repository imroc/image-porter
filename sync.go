package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
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
		sort.Sort(ImageTags(srcTags))

		if image.TagLimit != nil {
			limit := *image.TagLimit
			if limit < len(srcTags) {
				srcTags = srcTags[len(srcTags)-limit:]
			}
		}
		klog.V(2).Infof("filterd tag list for image %s (%s): %v", image.From, image.TagFilter.String(), srcTags)
		dstTags, err := listTags(image.To)
		if err != nil {
			klog.Infof("dst image have no tags: %s (%v)", image.To, err)
		}
		toSync, toUpdate := tagsToSync(srcTags, dstTags)
		if len(toSync) > 0 {
			klog.Infof("sync new tags for image %s: %v", image.From, toSync)
			syncImageTags(image.From, image.To, toSync, retryAttempt)
		} else {
			klog.Infof("no new tags for image %s", image.From)
		}
		if len(toUpdate) > 0 {
			updateImageTags(image.From, image.To, toUpdate, retryAttempt)
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

func tagsToSync(srcTags, dstTags []string) (toSync, toUpdate []string) {
	if len(dstTags) == 0 {
		toSync = srcTags
		return
	}
	dstMap := sliceToMap(dstTags)
	for _, tag := range srcTags {
		if !dstMap[tag] {
			toSync = append(toSync, tag)
			continue
		}
		_, err := semver.NewVersion(tag) // ignore updates of sematic version tag
		if err != nil {                  // not sematic version, may update
			if strings.Contains(tag, "-") { // ignore updates of the version that contains "-"
				continue
			}
			toUpdate = append(toUpdate, tag)
		}
	}
	return
}

func sliceToMap[T comparable](slice []T) map[T]bool {
	m := make(map[T]bool)
	for _, v := range slice {
		m[v] = true
	}
	return m
}

func inspectDigest(image, tag string) (digest string, err error) {
	args := []string{"inspect", "--format", "{{.Digest}}", fmt.Sprintf("docker://%s:%s", image, tag)}
	cmd := exec.Command("skopeo", args...)
	digestBytes, err := cmd.Output()
	if err != nil {
		klog.Errorf("failed to inspect %s:%s", image, tag)
	}
	digest = string(digestBytes)
	return
}

func updateImageTags(imageFrom, imageTo string, tags []string, retryAttempt int) {
	for _, tag := range tags {
		digestFrom, err := inspectDigest(imageFrom, tag)
		if err != nil {
			continue
		}
		digestTo, err := inspectDigest(imageTo, tag)
		if err != nil {
			continue
		}
		if string(digestFrom) != string(digestTo) { // found tag update
			klog.Infof("found tag update for image %s:%s", imageFrom, tag)
			syncImageTag(imageFrom, imageTo, tag, retryAttempt)
		}
	}
}

func syncImageTag(imageFrom, imageTo, tag string, retryAttempt int) {
	args := []string{"copy", "--all", fmt.Sprintf("docker://%s:%s", imageFrom, tag), fmt.Sprintf("docker://%s:%s", imageTo, tag)}
	klog.Infof("execute command: skopeo %s", strings.Join(args, " "))
	cmd := exec.Command("skopeo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	var err error
	for i := -1; i < retryAttempt; i++ {
		err = cmd.Run()
		if err == nil {
			continue
		}
	}
}

func syncImageTags(imageFrom, imageTo string, tags []string, retryAttempt int) {
	for i := len(tags) - 1; i >= 0; i-- {
		syncImageTag(imageFrom, imageTo, tags[i], retryAttempt)
	}
}
