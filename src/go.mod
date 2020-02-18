module github.com/src

require (
	github.com/hashicorp/consul v1.2.3
	github.com/labstack/gommon v0.2.9
	github.com/ystia/yorc/v4 v4.0.0-M5.0.20191031135939-0ba3732937aa
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
)

// Due to this capital letter thing we have troubles and we have to replace it explicitly
replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2

// Here you need to specify your local Yorc sources path
replace github.com/ystia/yorc/v4 v4.0.0-M5.0.20191031135939-0ba3732937aa => /my-path/to/yorc/github.com/ystia/yorc

go 1.13
