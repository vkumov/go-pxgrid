release:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin master v$(VERSION)
	curl  https://proxy.golang.org/github.com/vkumov/go-pxgrid/@v/v$(VERSION).info