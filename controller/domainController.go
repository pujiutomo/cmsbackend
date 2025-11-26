package controller

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/pujiutomo/cmsbackend/database"
	"github.com/pujiutomo/cmsbackend/models"
	"github.com/pujiutomo/cmsbackend/util"
	"gorm.io/gorm"
)

// constants
const (
	DomainKeyPattern = "domain:%d"
	DefaultPageLimit = 5
)

// struct data dari form
type DataRequest struct {
	ID          interface{} `json:"id"`
	Name        string      `json:"name"`
	Logo        string      `json:"logo"`
	MetaTitle   string      `json:"meta_title"`
	MetaDesc    string      `json:"meta_desc"`
	MetaKeyword string      `json:"meta_keyword"`
	MetaIco     string      `json:"meta_ico"`
	Modul       string      `json:"modul"`
	Status      string      `json:"status"`
	Aksi        string      `json:"aksi"`
}

// fungsi check duplikat nama domain
func checkDomainExists(name string, excludeID ...uint) (bool, error) {
	var domain models.Domain
	query := database.DB.Where("name = ?", strings.TrimSpace(name))

	if len(excludeID) > 0 && excludeID[0] != 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	result := query.First(&domain)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, result.Error
	}
	return domain.Id != 0, nil
}
func getDomainByID(id uint) (*models.Domain, error) {
	var domain models.Domain
	result := database.DB.Where("id = ?", id).First(&domain)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("domain with id %d not found", id)
		}
		return nil, result.Error
	}
	return &domain, nil
}

// fungsi persiapan data redis
func prepareDomainRedisData(domain models.Domain) (string, error) {
	tempMap := map[string]interface{}{
		"id":     domain.Id,
		"name":   domain.Name,
		"logo":   domain.Logo,
		"title":  domain.MetaTitle,
		"status": domain.Status,
	}

	//proses maping module
	var moduleMap []map[string]interface{}
	if domain.Modul != "" {
		arrayModul := strings.Split(domain.Modul, ",")
		for _, key := range arrayModul {
			trimmedKey := strings.TrimSpace(key)
			if trimmedKey != "" {
				newMap := util.ModulDesc(trimmedKey)
				moduleMap = append(moduleMap, newMap)
			}
		}
	}
	tempMap["modul"] = moduleMap

	jsonData, err := jsoniter.Marshal(tempMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal domain data: %w", err)
	}

	return string(jsonData), nil
}

// insert data ke database
func PostDomain(c *fiber.Ctx) error {
	var request DataRequest
	//formLogo := c.FormFile("logo")
	//formIco := c.FormFile("ico")

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
	}

	//validasi field
	validator := util.NewValidator().
		Required(request.Name, "name").
		Domain(request.Name, "name").
		Required(request.MetaTitle, "meta_title")

	if validator.HasErrors() {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation field",
			"errors":  validator.Errors(),
		})
	}

	//jika image exists
	namaLogo := ""
	fileLogo, err := c.FormFile("logo")
	if err == nil && fileLogo != nil && fileLogo.Filename != "" {
		validator := util.NewValidator().TypeImage(fileLogo, "logo").FileSize(fileLogo, "logo")
		if validator.HasErrors() {
			return c.Status(400).JSON(fiber.Map{
				"message": "validation field",
				"errors":  validator.Errors(),
			})
		}
		nmLogo, err := uploadController(fileLogo, "image")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": "field upload Logo",
				"errors":  err,
			})
		}
		namaLogo = nmLogo
	}

	namaIco := ""
	fileIco, err := c.FormFile("meta_ico")
	if err == nil && fileIco != nil && fileIco.Filename != "" {
		validator := util.NewValidator().TypeImage(fileIco, "meta_ico").FileSize(fileIco, "meta_ico")
		if validator.HasErrors() {
			return c.Status(400).JSON(fiber.Map{
				"message": "validation field",
				"errors":  validator.Errors(),
			})
		}
		nmIco, err := uploadController(fileIco, "image")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": "field upload Ico",
				"errors":  err,
			})
		}
		namaIco = nmIco
	}

	//check if domain exists
	exists, err := checkDomainExists(request.Name)
	if err != nil {
		log.Printf("Error checking domain existance: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error checking domain existance",
			"error":   err.Error(),
		})
	}
	if exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Domain already exists",
		})
	}

	//create domain object
	data := models.Domain{
		Name:        strings.TrimSpace(request.Name),
		Logo:        namaLogo,
		MetaTitle:   request.MetaTitle,
		MetaDesc:    request.MetaDesc,
		MetaKeyword: request.MetaKeyword,
		MetaIco:     namaIco,
		Modul:       request.Modul,
		Status:      request.Status,
	}

	//save to database
	if err := database.DB.Create(&data).Error; err != nil {
		log.Printf("Error creating domain: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Failed create domain",
			"error":   err.Error(),
		})
	}

	//prepare and save redis
	domainData, err := prepareDomainRedisData(data)
	if err != nil {
		log.Printf("warning: failed to prepare domain data for redis: %v", err)
	} else {
		domainKeyRedis := fmt.Sprintf(DomainKeyPattern, data.Id)
		if err := SaveToRedis(domainKeyRedis, domainData); err != nil {
			log.Printf("warning: failed to save data domain to redis: %v", err)
		}
	}

	return c.JSON(fiber.Map{
		"message": "Successfully saved data",
		"data": fiber.Map{
			"id": data.Id,
		},
	})
}

func UpdateDomain(c *fiber.Ctx) error {
	var requestData map[string]interface{}
	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
	}
	// Debug log untuk melihat data yang diterima
	log.Printf("UpdateDomain request data: %+v", requestData)

	aksi, ok := requestData["aksi"].(string)
	if !ok || aksi == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Action (aksi) is required",
		})
	}

	switch aksi {
	case "updateAll":
		return updateDomainAll(c, requestData)
	case "updateStatus":
		return updateDomainStatus(c, requestData)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid action specified: " + aksi,
		})
	}
}

func updateDomainStatus(c *fiber.Ctx, requestData map[string]interface{}) error {
	// Validasi data
	data, ok := requestData["data"].([]interface{})
	if !ok || len(data) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "No data provided",
		})
	}

	// Debug log untuk melihat data status update
	log.Printf("Status update data: %+v", data)

	var successCount int
	var errors []string

	// Process each item
	for i, item := range data {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("item %d: invalid data format", i))
			continue
		}

		// Extract dan convert ID
		itemID, ok := itemMap["id"]
		if !ok {
			errors = append(errors, fmt.Sprintf("item %d: id is required", i))
			continue
		}

		domainID, err := util.ConvertToUint(itemID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("item %d: %v", i, err))
			continue
		}

		// Extract status
		itemStatus, ok := itemMap["status"].(string)
		if !ok || itemStatus == "" {
			errors = append(errors, fmt.Sprintf("item %d: status is required and must be string", i))
			continue
		}

		// Check if domain exists
		_, err = getDomainByID(domainID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("item %d: %v", i, err))
			continue
		}

		// Update status
		result := database.DB.Model(&models.Domain{}).Where("id = ?", domainID).Update("status", itemStatus)
		if result.Error != nil {
			errors = append(errors, fmt.Sprintf("item %d: failed to update - %v", i, result.Error))
			continue
		}
		//fungsi update tapi tidak ada perubahan
		/*if result.RowsAffected == 0 {
			errors = append(errors, fmt.Sprintf("item %d: no rows affected", i))
			continue
		}*/
		// Get updated domain data for Redis
		updatedDomain, err := getDomainByID(domainID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("item %d: failed to fetch updated data - %v", i, err))
			continue
		}

		// Update Redis
		domainData, err := prepareDomainRedisData(*updatedDomain)
		if err != nil {
			log.Printf("Warning: Failed to prepare domain data for Redis (ID: %d): %v", domainID, err)
			// Continue meskipun Redis gagal, karena database update sudah berhasil
		} else {
			domainKeyRedis := fmt.Sprintf(DomainKeyPattern, updatedDomain.Id)
			if err := SaveToRedis(domainKeyRedis, domainData); err != nil {
				log.Printf("Warning: Failed to save domain to Redis (ID: %d): %v", domainID, err)
				// Continue meskipun Redis gagal
			}
		}

		successCount++
	}

	response := fiber.Map{
		"message": fmt.Sprintf("Successfully updated %d domains", successCount),
		"updated": successCount,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["message"] = fmt.Sprintf("Partially updated: %d success, %d failed", successCount, len(errors))

		if successCount == 0 {
			return c.Status(fiber.StatusInternalServerError).JSON(response)
		}

		return c.Status(fiber.StatusMultiStatus).JSON(response)
	}

	return c.JSON(response)
}

func updateDomainAll(c *fiber.Ctx, requestData map[string]interface{}) error {
	//extract dan conver id
	rawID, ok := requestData["id"]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ID is required",
		})
	}

	domainID, err := util.ConvertToUint(rawID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID format",
			"error":   err.Error(),
		})
	}

	//check if domain exists
	_, err = getDomainByID(domainID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	validator := util.NewValidator().
		Required(requestData["name"].(string), "name").
		Domain(requestData["name"].(string), "name").
		Required(requestData["meta_title"].(string), "meta_title")

	if validator.HasErrors() {
		return c.Status(400).JSON(fiber.Map{
			"message": "validation field",
			"errors":  validator.Errors(),
		})
	}

	// Validate name if provided
	if name, ok := requestData["name"].(string); ok && name != "" {
		// Check if domain name already exists (excluding current domain)
		exists, err := checkDomainExists(name, domainID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error checking domain existence",
				"error":   err.Error(),
			})
		}
		if exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Domain name already exists",
			})
		}
	}

	// Prepare update data
	updateData := map[string]interface{}{}

	if name, ok := requestData["name"].(string); ok {
		updateData["name"] = strings.TrimSpace(name)
	}
	if logo, ok := requestData["logo"].(string); ok {
		updateData["logo"] = logo
	}
	if metaTitle, ok := requestData["meta_title"].(string); ok {
		updateData["meta_title"] = metaTitle
	}
	if metaDesc, ok := requestData["meta_desc"].(string); ok {
		updateData["meta_desc"] = metaDesc
	}
	if metaKeyword, ok := requestData["meta_keyword"].(string); ok {
		updateData["meta_keyword"] = metaKeyword
	}
	if metaIco, ok := requestData["meta_ico"].(string); ok {
		updateData["meta_ico"] = metaIco
	}
	if modul, ok := requestData["modul"].(string); ok {
		updateData["modul"] = modul
	}
	if status, ok := requestData["status"].(string); ok {
		updateData["status"] = status
	}

	// Update domain
	result := database.DB.Model(&models.Domain{}).Where("id = ?", domainID).Updates(updateData)
	if result.Error != nil {
		log.Printf("Error updating domain: %v", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update domain",
			"error":   result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Domain not found or no changes made",
		})
	}

	// Get updated domain
	updatedDomain, err := getDomainByID(domainID)
	if err != nil {
		log.Printf("Warning: Failed to fetch updated domain: %v", err)
	} else {
		// Update Redis
		domainData, err := prepareDomainRedisData(*updatedDomain)
		if err != nil {
			log.Printf("Warning: Failed to prepare domain data for Redis: %v", err)
		} else {
			domainKeyRedis := fmt.Sprintf(DomainKeyPattern, updatedDomain.Id)
			if err := SaveToRedis(domainKeyRedis, domainData); err != nil {
				log.Printf("Warning: Failed to save domain to Redis: %v", err)
			}
		}
	}

	return c.JSON(fiber.Map{
		"message": "Successfully updated domain",
		"data": fiber.Map{
			"id": domainID,
		},
	})
}

func GetDomain(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", strconv.Itoa(DefaultPageLimit)))
	if err != nil || limit < 1 {
		limit = DefaultPageLimit
	}

	// Set maximum limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	var domains []models.Domain
	var total int64

	// Get total count
	if err := database.DB.Model(&models.Domain{}).Count(&total).Error; err != nil {
		log.Printf("Error counting domains: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to count domains",
			"error":   err.Error(),
		})
	}

	// Get paginated data
	if err := database.DB.Offset(offset).Limit(limit).Order("id DESC").Find(&domains).Error; err != nil {
		log.Printf("Error fetching domains: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch domains",
			"error":   err.Error(),
		})
	}

	lastPage := int(math.Ceil(float64(total) / float64(limit)))
	if lastPage < 1 {
		lastPage = 1
	}

	return c.JSON(fiber.Map{
		"data": domains,
		"meta": fiber.Map{
			"total":     total,
			"page":      page,
			"limit":     limit,
			"last_page": lastPage,
			"from":      offset + 1,
			"to":        offset + len(domains),
		},
	})
}
