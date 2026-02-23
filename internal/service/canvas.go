package service

import (
	"WorksKeeper/internal/entity"
	"WorksKeeper/internal/frontend"
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/view"
)

type CanvasService struct {
	canvasRepo  *repository.CanvasRepository
	captionRepo *repository.CaptionRepository
	groupRepo   *repository.GroupRepository
	mediaRepo   *repository.MediaRepository
	sourceRepo  *repository.SourceRepository
	textRepo    *repository.TextRepository
	workRepo    *repository.WorkRepository
}

func NewCanvasService(
	canvasRepo *repository.CanvasRepository,
	captionRepo *repository.CaptionRepository,
	groupRepo *repository.GroupRepository,
	mediaRepo *repository.MediaRepository,
	sourceRepo *repository.SourceRepository,
	textRepo *repository.TextRepository,
	workRepo *repository.WorkRepository) *CanvasService {
	return &CanvasService{
		canvasRepo:  canvasRepo,
		captionRepo: captionRepo,
		groupRepo:   groupRepo,
		mediaRepo:   mediaRepo,
		sourceRepo:  sourceRepo,
		textRepo:    textRepo,
		workRepo:    workRepo,
	}
}

/*func (cs *CanvasService) GetWork(id int64) frontend.Canvasable {
	return cs.workRepo.GetWork(id)
}*/

// TODO: needs to create a canvas and root group
func (cs *CanvasService) GetNewWorkView() *view.WorkView {
	return view.NewEmptyWorkView(cs.workRepo.InsertNewWork())
}

func (cs *CanvasService) GetWorkView(id int64) *view.WorkView {
	// Get single entity.Work
	// work := GetWork
	work := cs.workRepo.GetWork(id)

	// Get single entity.Canvas
	// canvas := GetCanvasByWork(work)
	canvas := cs.canvasRepo.GetCanvasByWork(work)

	// Get []entity.Group for canvas
	// groups := GetGroupsByCanvas(canvas)
	groups := cs.groupRepo.GetGroupsByCanvas(canvas)

	// Get map[int64][]entity.Media for []entity.Group
	// media := GetMediasByGroupsOrderedByGroupIdAndIdx(groups)
	media := cs.mediaRepo.GetMediasByGroupsOrderedByGroupIdAndIdx(groups)

	// Get map[int64][]entity.Text for []entity.Group
	// texts := GetTextsByGroupsOrderedByGroupIdAndIdx(groups)
	texts := cs.textRepo.GetTextsByGroupsOrderedByGroupIdAndIdx(groups)

	var flatMedia []*entity.Media
	for _, m := range media {
		flatMedia = append(flatMedia, m...)
	}

	// Get map[int64][]entity.Source for []entity.Media (maps.Values(map[int64][]entity.Media))
	// sources := GetSourcesByMedias(maps.Values(media))
	sources := cs.sourceRepo.GetSourcesByMedias(flatMedia)

	// Get map[int64]entity.Caption for []entity.Media
	// captions := GetCaptionByMedias(maps.Values(media))
	captions := cs.captionRepo.GetCaptionByMedias(flatMedia)

	// COMPOSE VIEWS

	groupViewMap := make(map[int64]*view.GroupView)
	for _, g := range groups {

		groupMedia := media[g.Id]
		groupTexts := texts[g.Id]

		// Merge cursors (slice positions)
		mi, ti := 0, 0

		// Next expected global child index
		currIdx := int32(0)

		// Lower bound: media + texts
		groupChildren := make(
			[]frontend.Templatable,
			len(groupMedia)+len(groupTexts),
		)

		for mi < len(groupMedia) || ti < len(groupTexts) {

			var nextIsMedia bool

			if ti >= len(groupTexts) {
				nextIsMedia = true
			} else if mi >= len(groupMedia) {
				nextIsMedia = false
			} else {
				nextIsMedia = groupMedia[mi].Idx < groupTexts[ti].Idx
			}

			if nextIsMedia {
				m := groupMedia[mi]

				// Gap implies groups
				if m.Idx > currIdx {
					nbrGroups := m.Idx - currIdx
					groupChildren = append(
						groupChildren,
						make([]frontend.Templatable, nbrGroups)...,
					)
					currIdx += nbrGroups
				}

				// Build media view
				var sourceViews []*view.SourceView
				for _, s := range sources[m.Id] {
					sourceViews = append(
						sourceViews,
						view.NewSourceView(s),
					)
				}

				captionView := view.NewCaptionView(captions[m.Id])
				groupChildren[m.Idx] = view.NewMediaView(m, sourceViews, captionView)

				mi++
				currIdx++

			} else {
				t := groupTexts[ti]

				// Gap implies groups
				if t.Idx > currIdx {
					nbrGroups := t.Idx - currIdx
					groupChildren = append(
						groupChildren,
						make([]frontend.Templatable, nbrGroups)...,
					)
					currIdx += nbrGroups
				}

				groupChildren[t.Idx] = view.NewTextView(t)

				ti++
				currIdx++
			}
		}

		groupView := view.NewGroupView(g, groupChildren)
		groupViewMap[g.Id] = groupView
	}

	var rootGroup *view.GroupView
	for _, g := range groups {
		groupView := groupViewMap[g.Id]

		if g.ParentId == nil {
			rootGroup = groupView
		} else {
			parentGroup := groupViewMap[*g.ParentId]
			parentGroup.PutChild(groupView, *g.Idx)
		}
	}

	canvasView := view.NewCanvasView(canvas, rootGroup)
	workView := view.NewWorkView(work, canvasView)

	return workView
}
