package router

import (
	"fynance/helpers"
	"fynance/views"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Router struct {
	Window fyne.Window
	UserId primitive.ObjectID
}

func NewRouter(window fyne.Window) *Router {
	return &Router{
		Window: window,
		UserId: helpers.CurrentUserID,
	}
}

func (r *Router) layout(content fyne.CanvasObject) {

	sidebar := views.Sidebar(
		r.Window,
		r.ShowParameters,
		r.ShowIncome,
		r.ShowExpenses,
		r.ShowDailyReport,
		r.ShowWeeklyReport,
		r.ShowMonthlyReport,
		r.ShowContact,
		r.ShowDashboard,
		r.ShowLogin,
		r.UserId,
	)

	r.Window.SetContent(
		container.NewBorder(nil, nil, sidebar, nil, content),
	)
}

func (r *Router) ShowDashboard() {
	r.layout(views.Dashboard(r.Window))
}

func (r *Router) ShowDailyReport() {
	r.layout(views.DailyReport(r.Window))
}

func (r *Router) ShowWeeklyReport() {
	r.layout(views.WeeklyReport(r.Window))
}

func (r *Router) ShowIncome() {
	r.layout(views.IncomeView(r.Window, r.UserId))
}

func (r *Router) ShowMonthlyReport() {
	r.layout(views.MonthlyReport(r.Window))
}

func (r *Router) ShowExpenses() {
	r.layout(views.ExpenseView(r.Window, r.UserId))
}

func (r *Router) ShowParameters() {
	r.layout(views.ParametersView(r.Window, r.UserId))
}

func (r *Router) ShowContact() {
	r.layout(views.ContactView(r.Window))
}

func (r *Router) ShowLogin() {
	r.Window.SetContent(
		views.LoginView(
			r.Window,
			r.ShowDashboard,
		),
	)
}
