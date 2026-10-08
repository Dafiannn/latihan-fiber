package utils

import "github.com/gofiber/fiber/v2"

// SuccessResponse adalah format standar untuk HTTP 200/201 sesuai PDF
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}, meta interface{}) error {
	res := fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	}
	if meta != nil {
		res["meta"] = meta
	}
	return c.Status(statusCode).JSON(res)
}

// ErrorResponse adalah format standar untuk HTTP Error (401, 403, 404, 422, dsb)
func ErrorResponse(c *fiber.Ctx, statusCode int, message string, errors interface{}) error {
	res := fiber.Map{
		"success": false,
		"message": message,
	}
	if errors != nil {
		res["errors"] = errors
	}
	return c.Status(statusCode).JSON(res)
}
