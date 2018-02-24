package apis

import (
	"fmt"
	"gofreta/app"
	"gofreta/app/daos"
	"gofreta/app/models"
	"gofreta/app/utils"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	routing "github.com/go-ozzo/ozzo-routing"
)

// MediaApi defines media api services
type MediaApi struct {
	routeGroup   *routing.RouteGroup
	mongoSession *mgo.Session
	dao          *daos.MediaDAO
}

// InitMediaApi sets up the routing of media endpoints and the corresponding handlers.
func InitMediaApi(rg *routing.RouteGroup, session *mgo.Session) {
	api := MediaApi{
		routeGroup:   rg,
		mongoSession: session,
		dao:          daos.NewMediaDAO(session),
	}

	rg.Get("/media", authenticateToken(session, "media", "index"), usersOnly, api.index)
	rg.Post("/media", authenticateToken(session, "media", "upload"), usersOnly, api.upload)
	rg.Get("/media/<id>", authenticateToken(session, "media", "view"), usersOnly, api.view)
	rg.Put("/media/<id>", authenticateToken(session, "media", "update"), usersOnly, api.update)
	rg.Post("/media/<id>", authenticateToken(session, "media", "replace"), usersOnly, api.replace)
	rg.Delete("/media/<id>", authenticateToken(session, "media", "delete"), usersOnly, api.delete)
}

// index api handler for fetching paginated media items list
func (api *MediaApi) index(c *routing.Context) error {
	// --- fetch search data
	searchFields := []string{"title", "type", "path", "created", "modified"}
	searchData := utils.GetSearchConditions(c, searchFields)
	// ---

	// --- fetch sort data
	sortFields := []string{"title", "type", "path", "created", "modified"}
	sortData := utils.GetSortFields(c, sortFields)
	// ---

	total, _ := api.dao.Count(searchData)

	limit, page := utils.GetPaginationSettings(c, total)

	utils.SetPaginationHeaders(c, limit, total, page)

	items := []models.Media{}

	if total > 0 {
		items, _ = api.dao.GetList(limit, limit*(page-1), searchData, sortData)

		items = daos.ToAbsMediaPaths(items)
	}

	return c.Write(items)
}

// view api handler for fetching single media item data
func (api *MediaApi) view(c *routing.Context) error {
	id := c.Param("id")

	model, err := api.dao.GetByID(id)

	if err != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Media item with id \"%v\" doesn't exist!", id))
	}

	model = daos.ToAbsMediaPath(model)

	return c.Write(model)
}

// update api handler for updating existing media item settings (eg. name)
func (api *MediaApi) update(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Media item with id \"%v\" doesn't exist!", id))
	}

	form := &models.MediaUpdateForm{}
	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating media item.", readErr)
	}

	form.Model = model

	updatedModel, updateErr := api.dao.Update(form)

	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while updating media item.", updateErr)
	}

	updatedModel = daos.ToAbsMediaPath(updatedModel)

	return c.Write(updatedModel)
}

// delete api handler for deleting an existing media item
func (api *MediaApi) delete(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Media item with id \"%v\" doesn't exist!", id))
	}

	deleteErr := api.dao.Delete(model)
	if deleteErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while deleting Language item.", deleteErr)
	}

	c.Response.WriteHeader(http.StatusNoContent)

	return nil
}

// upload api handler for uploading and creating multiple media items
func (api *MediaApi) upload(c *routing.Context) error {
	items := []*models.Media{}
	errs := []error{}

	//get the multipart reader for the request.
	reader, readerErr := c.Request.MultipartReader()
	if readerErr != nil {
		return readerErr
	}

	// copy each part to destination
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		// if no file name is provided, skip the iteration
		if part.FileName() == "" {
			continue
		}

		uploadPath, uploadFileType, uploadErr := uploadFile(part)
		if uploadErr != nil {
			errs = append(errs, uploadErr)

			continue
		}

		name := part.FileName()
		ext := filepath.Ext(name)
		basename := strings.TrimSuffix(name, ext)

		mediaData := &models.Media{
			Title: basename,
			Path:  uploadPath,
			Type:  uploadFileType,
		}

		// db record create
		model, createErr := api.dao.Create(mediaData)
		if createErr != nil {
			errs = append(errs, createErr)
			continue
		}

		model = daos.ToAbsMediaPath(model)

		items = append(items, model)
	}

	data := &struct {
		Items  []*models.Media `json:"items"`
		Errors []error         `json:"errors"`
	}{items, errs}

	return c.Write(data)
}

// replace api handler for replacing existing media file
func (api *MediaApi) replace(c *routing.Context) error {
	id := c.Param("id")

	model, fetchErr := api.dao.GetByID(id)
	if fetchErr != nil {
		return utils.NewNotFoundError(fmt.Sprintf("Media item with id \"%v\" doesn't exist!", id))
	}

	file, _, readErr := c.Request.FormFile("file")
	if readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while replacing media file.", readErr)
	}
	defer file.Close()

	filePath, fileType, uploadErr := uploadFile(file)
	if uploadErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while replacing media file.", uploadErr)
	}

	model.Path = filePath
	model.Type = fileType

	// db record update
	updatedModel, updateErr := api.dao.Replace(model)
	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while replacing media file.", updateErr)
	}

	updatedModel = daos.ToAbsMediaPath(updatedModel)

	return c.Write(updatedModel)
}

// uploadFile validates and uploads single file from a reader (returns the uploaded file path and type on success).
func uploadFile(r io.Reader) (string, string, error) {
	b, readErr := ioutil.ReadAll(r)
	if readErr != nil {
		return "", "", readErr
	}

	// --- validate
	isValid := utils.ValidateMimeType(b, models.ValidMediaTypes())
	if isValid != true {
		return "", "", utils.NewDataError("Invalid or unsupported file type.")
	}

	sizeErr := utils.ValidateSize(b, gofreta.App.Config.GetFloat64("upload.maxSize"))
	if sizeErr != true {
		return "", "", utils.NewDataError("Media is too big.")
	}
	// ---

	// --- normalize
	ext, fileType := utils.GetExtAndFileTypeByMimeType(b)
	name := utils.MD5(utils.Random(10)+time.Now().String()) + "." + ext

	uploadDir := strings.TrimSuffix(gofreta.App.Config.GetString("upload.dir"), "/")
	uploadPath := uploadDir + "/" + name

	// ensure that the upload directory exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0777)
	}
	// ---

	// save file
	dst, dstCreateErr := os.Create(uploadPath)
	if dstCreateErr != nil {
		return "", "", dstCreateErr
	}
	defer dst.Close()

	_, writeErr := dst.Write(b)
	if writeErr != nil {
		return "", "", writeErr
	}

	return uploadPath, fileType, nil
}
