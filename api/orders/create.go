package orders

import (
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
)

func (orm *OrderRoutesManager) CreateOrder(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.OrderRequest](r)
	if err != nil {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidRequestBody"),
			gecho.WithData(err),
			gecho.Send(),
		)
		return
	}

	if body.Email == "" || body.Name == "" || len(body.Products) == 0 {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.missingFields"),
			gecho.Send(),
		)
		return
	}

	orderNum := orm.orderService.GenerateOrderNumber()

	orderId := uuid.New()

	order := &tables.Order{
		Id:          orderId,
		Email:       body.Email,
		Name:        body.Name,
		OrderNumber: orderNum,
	}

	err = orm.orderService.CreateOrder(r.Context(), order)
	if err != nil {
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.creatingOrder"),
			gecho.WithData(err),
			gecho.Send(),
		)
		return
	}

	orderLines := make([]*tables.OrderLine, len(body.Products))
	for id, quantity := range body.Products {
		uuid, err := uuid.Parse(id)
		if err != nil {
			gecho.BadRequest(w,
				gecho.WithMessage("error.order.invalidProductID"),
				gecho.WithData(err),
				gecho.Send(),
			)
			return
		}
		orderLines = append(orderLines, &tables.OrderLine{
			OrderId:   orderId,
			ProductId: uuid,
			Quantity:  quantity,
		})
	}

	err = orm.orderService.CreateOrderLines(r.Context(), orderLines)
	if err != nil {
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.creatingOrderLines"),
			gecho.WithData(err),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.order.created"),
		gecho.WithData(map[string]string{
			"order_number": orderNum,
		}),
		gecho.Send(),
	)
}
