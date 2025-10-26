package ecr

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jaehong21/hibiscus/internal/aws/ecr"
	"github.com/jaehong21/hibiscus/tviewapp/hibiscus"
	"github.com/jaehong21/hibiscus/utils"
)

type tab int

const (
	repoTab tab = iota
	imageTab
)

// Service implements the hibiscus.Service interface for Amazon ECR.
type Service struct {
	ctx hibiscus.ServiceContext

	layout     *tview.Flex
	pages      *tview.Pages
	filter     *tview.InputField
	repoTable  *tview.Table
	imageTable *tview.Table

	current tab

	repos          []types.Repository
	filteredRepos  []types.Repository
	images         []types.ImageDetail
	filteredImages []types.ImageDetail
	currentRepo    string
	currentRepoURI string

	mu     sync.Mutex
	active bool
}

func New(ctx hibiscus.ServiceContext) hibiscus.Service {
	svc := &Service{ctx: ctx, current: repoTab}
	svc.filter = tview.NewInputField().
		SetLabel("Filter (/): ").
		SetFieldBackgroundColor(tcell.ColorBlack)

	svc.repoTable = buildTable("ECR repositories")
	svc.imageTable = buildTable("Repository images")

	svc.pages = tview.NewPages()
	svc.pages.AddPage("repos", svc.repoTable, true, true)
	svc.pages.AddPage("images", svc.imageTable, true, false)

	svc.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(svc.filter, 1, 0, true).
		AddItem(svc.pages, 0, 1, true)

	svc.filter.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			svc.applyFilter(strings.TrimSpace(svc.filter.GetText()))
		case tcell.KeyEsc:
			svc.exitFilterMode()
		}
	})

	return svc
}

func (s *Service) Name() string  { return "ecr" }
func (s *Service) Title() string { return "Amazon ECR – repositories › images" }
func (s *Service) Primitive() tview.Primitive {
	return s.layout
}

func (s *Service) Init() {
	s.loadRepos()
}

func (s *Service) Activate() {
	s.active = true
	s.focusCurrentTable()
}

func (s *Service) Deactivate() {
	s.active = false
}

func (s *Service) Refresh() {
	if s.current == imageTab && s.currentRepo != "" {
		s.loadImages(s.currentRepo)
		return
	}
	s.loadRepos()
}

func (s *Service) EnterFilterMode() bool {
	if !s.canFocus() {
		return false
	}
	s.ctx.App.SetFocus(s.filter)
	return true
}

func (s *Service) InFilterMode() bool {
	return s.filter.HasFocus()
}

func (s *Service) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if event == nil {
		return nil
	}

	switch event.Key() {
	case tcell.KeyEsc:
		if s.filter.HasFocus() {
			s.exitFilterMode()
			return nil
		}
		if s.current == imageTab {
			s.showRepoTab()
			return nil
		}
	case tcell.KeyEnter:
		if s.repoTable.HasFocus() {
			s.openSelectedRepository()
			return nil
		}
	}

	switch event.Rune() {
	case 'c', 'C', 'y', 'Y':
		if s.repoTable.HasFocus() {
			s.copySelectedRepo()
			return nil
		}
		if s.imageTable.HasFocus() {
			s.copySelectedImage()
			return nil
		}
	}

	return event
}

func (s *Service) exitFilterMode() {
	s.filter.SetText("")
	switch s.current {
	case imageTab:
		s.setFocus(s.imageTable)
	default:
		s.setFocus(s.repoTable)
	}
}

func (s *Service) openSelectedRepository() {
	row, _ := s.repoTable.GetSelection()
	if row <= 0 || row-1 >= len(s.filteredRepos) {
		return
	}
	repo := s.filteredRepos[row-1]
	if repo.RepositoryName == nil {
		return
	}
	s.currentRepo = *repo.RepositoryName
	s.currentRepoURI = valueOr(repo.RepositoryUri)
	s.loadImages(s.currentRepo)
}

func (s *Service) copySelectedRepo() {
	row, _ := s.repoTable.GetSelection()
	if row <= 0 || row-1 >= len(s.filteredRepos) {
		return
	}
	repo := s.filteredRepos[row-1]
	if repo.RepositoryUri == nil {
		return
	}
	if err := clipboard.WriteAll(*repo.RepositoryUri); err != nil {
		s.ctx.SetError(fmt.Errorf("failed to copy URI: %w", err))
		return
	}
	s.ctx.SetError(nil)
	s.ctx.SetStatus("Repository URI copied to clipboard")
}

func (s *Service) copySelectedImage() {
	if s.currentRepo == "" {
		return
	}
	row, _ := s.imageTable.GetSelection()
	if row <= 0 || row-1 >= len(s.filteredImages) {
		return
	}
	image := s.filteredImages[row-1]
	if len(image.ImageTags) == 0 || image.ImageTags[0] == "" {
		return
	}

	tag := image.ImageTags[0]
	repoURI := s.currentRepoURI
	if repoURI == "" {
		repoURI = s.currentRepo
	}
	repoURI = fmt.Sprintf("%s:%s", repoURI, tag)
	if err := clipboard.WriteAll(repoURI); err != nil {
		s.ctx.SetError(fmt.Errorf("failed to copy image URI: %w", err))
		return
	}
	s.ctx.SetError(nil)
	s.ctx.SetStatus("Image URI copied to clipboard")
}

func (s *Service) loadRepos() {
	s.ctx.SetStatus("Fetching repositories...")
	s.ctx.SetError(nil)

	go func() {
		repos, err := ecr.DescribeRepositories()
		s.ctx.App.QueueUpdateDraw(func() {
			if err != nil {
				s.ctx.SetError(fmt.Errorf("describe repositories: %w", err))
				return
			}
			s.mu.Lock()
			s.repos = repos
			s.filteredRepos = append([]types.Repository(nil), repos...)
			s.mu.Unlock()
			s.renderRepos()
			s.showRepoTab()
			s.ctx.SetStatus(fmt.Sprintf("Loaded %d repositories", len(repos)))
		})
	}()
}

func (s *Service) loadImages(repo string) {
	if repo == "" {
		return
	}
	s.ctx.SetStatus(fmt.Sprintf("Fetching images for %s...", repo))
	s.ctx.SetError(nil)

	repoName := repo
	go func() {
		images, err := ecr.DescribeImages(&repoName)
		s.ctx.App.QueueUpdateDraw(func() {
			if err != nil {
				s.ctx.SetError(fmt.Errorf("describe images: %w", err))
				return
			}
			s.mu.Lock()
			s.images = images
			s.filteredImages = append([]types.ImageDetail(nil), images...)
			s.mu.Unlock()
			s.renderImages()
			s.showImageTab(repoName)
			s.ctx.SetStatus(fmt.Sprintf("Loaded %d images", len(images)))
		})
	}()
}

func (s *Service) applyFilter(query string) {
	s.ctx.SetError(nil)
	query = strings.ToLower(strings.TrimSpace(query))

	switch s.current {
	case imageTab:
		s.filteredImages = s.filteredImages[:0]
		if query == "" {
			s.filteredImages = append(s.filteredImages, s.images...)
		} else {
			for _, image := range s.images {
				if matchImage(image, query) {
					s.filteredImages = append(s.filteredImages, image)
				}
			}
		}
		s.renderImages()
	default:
		s.filteredRepos = s.filteredRepos[:0]
		if query == "" {
			s.filteredRepos = append(s.filteredRepos, s.repos...)
		} else {
			for _, repo := range s.repos {
				if repo.RepositoryName == nil {
					continue
				}
				if strings.Contains(strings.ToLower(*repo.RepositoryName), query) {
					s.filteredRepos = append(s.filteredRepos, repo)
				}
			}
		}
		s.renderRepos()
	}

	s.filter.SetText("")
	s.exitFilterMode()
}

func (s *Service) renderRepos() {
	table := s.repoTable
	table.Clear()

	headers := []string{"Repository", "URI", "Created"}
	for col, title := range headers {
		table.SetCell(0, col, headerCell(title))
	}

	if len(s.filteredRepos) == 0 {
		table.SetCell(1, 0, tableCell("No repositories found").SetSelectable(false))
		return
	}

	for idx, repo := range s.filteredRepos {
		name := valueOr(repo.RepositoryName)
		uri := valueOr(repo.RepositoryUri)
		created := ""
		if repo.CreatedAt != nil {
			created = repo.CreatedAt.Local().Format(time.RFC3339)
		}

		table.SetCell(idx+1, 0, tableCell(name))
		table.SetCell(idx+1, 1, tableCell(uri))
		table.SetCell(idx+1, 2, tableCell(created))
	}

	table.Select(1, 0)
}

func (s *Service) renderImages() {
	table := s.imageTable
	table.Clear()

	headers := []string{"Tag", "Pushed at", "Size", "Digest"}
	for col, title := range headers {
		table.SetCell(0, col, headerCell(title))
	}

	if len(s.filteredImages) == 0 {
		msg := "No images loaded"
		if s.currentRepo != "" {
			msg = "No images match this filter"
		}
		table.SetCell(1, 0, tableCell(msg))
		table.Select(1, 0)
		return
	}

	for idx, image := range s.filteredImages {
		tag := ""
		if len(image.ImageTags) > 0 {
			tag = image.ImageTags[0]
		}
		pushed := ""
		if image.ImagePushedAt != nil {
			pushed = image.ImagePushedAt.Local().Format(time.RFC3339)
		}
		size := utils.GetSizeFromByte(image.ImageSizeInBytes)
		digest := valueOr(image.ImageDigest)

		table.SetCell(idx+1, 0, tableCell(tag))
		table.SetCell(idx+1, 1, tableCell(pushed))
		table.SetCell(idx+1, 2, tableCell(size))
		table.SetCell(idx+1, 3, tableCell(digest))
	}

	table.Select(1, 0)
}

func (s *Service) showRepoTab() {
	s.current = repoTab
	s.pages.SwitchToPage("repos")
	s.repoTable.SetTitle("ECR repositories")
	s.setFocus(s.repoTable)
}

func (s *Service) showImageTab(repo string) {
	s.current = imageTab
	s.pages.SwitchToPage("images")
	s.imageTable.SetTitle(fmt.Sprintf("Images for %s", repo))
	s.setFocus(s.imageTable)
}

func matchImage(image types.ImageDetail, query string) bool {
	if len(image.ImageTags) > 0 {
		for _, tag := range image.ImageTags {
			if strings.Contains(strings.ToLower(tag), query) {
				return true
			}
		}
	}
	if image.ImageDigest != nil && strings.Contains(strings.ToLower(*image.ImageDigest), query) {
		return true
	}
	return false
}

func headerCell(title string) *tview.TableCell {
	return tview.NewTableCell(title).
		SetTextColor(tcell.ColorLightCyan).
		SetSelectable(false).
		SetAlign(tview.AlignLeft)
}

func tableCell(value string) *tview.TableCell {
	return tview.NewTableCell(value).
		SetExpansion(1)
}

func valueOr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func buildTable(title string) *tview.Table {
	tbl := tview.NewTable().
		SetFixed(1, 0).
		SetSelectable(true, false)
	tbl.SetBorder(true)
	tbl.SetTitle(title)
	tbl.SetBorderColor(tcell.ColorDimGray)
	return tbl
}

func (s *Service) canFocus() bool {
	return s.ctx.App != nil && s.active
}

func (s *Service) setFocus(p tview.Primitive) {
	if !s.canFocus() || p == nil {
		return
	}
	s.ctx.App.SetFocus(p)
}

func (s *Service) focusCurrentTable() {
	switch s.current {
	case imageTab:
		s.setFocus(s.imageTable)
	default:
		s.setFocus(s.repoTable)
	}
}
