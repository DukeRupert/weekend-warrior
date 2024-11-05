

## App flow
Login ->
  if (super) ->
    * Facility List 
    * Create facility form
    * Assign admin form
  if (admin) -> Redirect to facility
    * Dashboard calender
    * Modify all Schedule
    * Controller list 
    * Create Controller form 
    * Assign Schedule form
    * Assign Admin form
  if (user) -> Redirect to facility
    * Dashboard calender
    * Modify own Schedule
    * Update controller info form

## Rough router layout
```golang
```
  // Authentication routes
	auth := app.Group("/auth")
	{
		auth.Post("/login", handleLogin)
		auth.Post("/logout", handleLogout)
	}

	// Super admin routes
	superAdmin := app.Group("/super", authMiddleware, roleCheck("super"))
	{
		// Facility management
		superAdmin.Get("/facilities", listAllFacilities)
		superAdmin.Post("/facilities", createFacility)
		superAdmin.Get("/facilities/:id", getFacility)
		superAdmin.Put("/facilities/:id", updateFacility)
		superAdmin.Delete("/facilities/:id", deleteFacility)
		
		// Initial admin assignment
		superAdmin.Post("/facilities/:id/admin", assignInitialAdmin)
	}

	// Admin routes
	admin := app.Group("/admin", authMiddleware, roleCheck("admin"), facilityCheck)
	{
		// Facility management (limited to their facility)
		admin.Get("/facility", getAdminFacility)
		admin.Put("/facility", updateAdminFacility)

		// Controller management
		admin.Get("/controllers", listFacilityControllers)
		admin.Post("/controllers", createController)
		admin.Put("/controllers/:id", updateController)
		admin.Delete("/controllers/:id", deleteController)
		
		// Role management
		admin.Post("/controllers/:id/role", assignAdminRole)
		
		// Schedule management
		admin.Get("/schedules", listAllSchedules)
		admin.Post("/schedules", createSchedule)
		admin.Put("/schedules/:id", updateSchedule)
		
		// Availability management
		admin.Get("/availability", getFullAvailability)
		admin.Put("/availability/:controller_id", updateControllerAvailability)
	}

	// User routes
	user := app.Group("/user", authMiddleware, roleCheck("user"), facilityCheck)
	{
		// View own facility info
		user.Get("/facility", getUserFacility)
		
		// Schedule viewing
		user.Get("/schedule", getUserSchedule)
		
		// Availability management (only their own)
		user.Get("/availability", getUserAvailability)
		user.Put("/availability", updateUserAvailability)
		user.Post("/availability/toggle/:date", toggleAvailability)
	}
```
```
