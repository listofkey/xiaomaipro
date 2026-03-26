package test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/customer"
	"github.com/stripe/stripe-go/v84/price"
	"github.com/stripe/stripe-go/v84/product"
	"github.com/stripe/stripe-go/v84/webhook"
)

func TestPay(t *testing.T) {
	stripe.Key = stripeSecretKeyForTest(t)

	product_params := &stripe.ProductParams{
		Name:        stripe.String("Starter Subscription"),
		Description: stripe.String("$12/Month subscription"),
	}
	starter_product, _ := product.New(product_params)

	price_params := &stripe.PriceParams{
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Product:  stripe.String(starter_product.ID),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
		UnitAmount: stripe.Int64(1200),
	}
	starter_price, _ := price.New(price_params)

	fmt.Println("Success! Here is your starter subscription product id: " + starter_product.ID)
	fmt.Println("Success! Here is your starter subscription price id: " + starter_price.ID)
}

func TestRequest(t *testing.T) {
	stripe.Key = stripeSecretKeyForTest(t)
	params := &stripe.CustomerParams{
		Email:       stripe.String("jane.smith@email.com"),
		Name:        stripe.String("Jane Smith"),
		Description: stripe.String("My First Stripe Customer"),
	}
	result, err := customer.New(params)
	if err != nil {
		fmt.Printf("customer.New error: %+v\n", err)
		return
	}
	fmt.Printf("customer.New result: %+v\n", result)
}

func TestPayPrice(t *testing.T) {
	stripe.Key = stripeSecretKeyForTest(t)
	// 1. 构建收银台会话参数
	params := &stripe.CheckoutSessionParams{
		// 支付方式：可以写 card, alipay, wechat_pay 等
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}), 
		
		// 订单明细（Stripe 收银台页面会展示买的是什么）
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("你的高级会员服务"),
						Description: stripe.String("享受一年高级特权"),
					},
					UnitAmount: stripe.Int64(5000), // $50.00
				},
				Quantity: stripe.Int64(1), // 数量
			},
		},
		
		// 模式：payment (单次支付), setup (绑卡), subscription (订阅)
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)), 
		
		// 支付成功后跳回你 Vue 网站的哪个页面
		SuccessURL: stripe.String("https://bilibili.com/success?session_id={CHECKOUT_SESSION_ID}"),
		// 用户点取消跳回哪个页面
		CancelURL:  stripe.String("https://bilibili.com/cart"),
		
		// 同样需要挂载你的内部订单号，方便 Webhook 查单
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"my_internal_order_sn": "ORD-2026-0002",
			},
		},
	}
	// 2. 调用 Stripe API 创建 Session
	s, err := session.New(params)
	if err != nil {
		log.Fatalf("创建 Session 失败: %v", err)
	}

	// 3. 将这个 URL 返回给前端，前端直接打开这个链接
	fmt.Printf("让前端跳转到这个链接完成支付: %s\n", s.URL)
}

func TestSuccessful(t *testing.T) {
	webhookSecret := stripeWebhookSecretForTest(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body failed", http.StatusBadRequest)
			log.Printf("读取 webhook body 失败: %v", err)
			return
		}

		event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
		if err != nil {
			http.Error(w, "signature verify failed", http.StatusBadRequest)
			log.Printf("Stripe webhook 验签失败: %v", err)
			return
		}

		log.Printf("收到 Stripe webhook: type=%s id=%s", event.Type, event.ID)
		handleWebhookEvent(event)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8088"
	log.Printf("Stripe webhook listener started at http://127.0.0.1%s/webhook", addr)
	log.Printf("请使用 `stripe listen --forward-to http://127.0.0.1%s/webhook` 转发事件", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		t.Fatalf("webhook server failed: %v", err)
	}
}

func handleWebhookEvent(event stripe.Event) {
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("解析 Session 失败: %v", err)
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			orderSN := session.Metadata["my_internal_order_sn"]
			log.Printf("支付成功！内部订单号: %s, 实际支付金额: %d", orderSN, session.AmountTotal)
			return
		}

		log.Printf("checkout.session.completed 但未支付成功，payment_status=%s", session.PaymentStatus)

	case "checkout.session.expired":
		log.Println("订单已过期，可以释放库存")

	default:
		log.Printf("忽略未处理事件: %s", event.Type)
	}
}

func stripeSecretKeyForTest(t *testing.T) string {
	t.Helper()

	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		t.Skip("STRIPE_SECRET_KEY is empty")
	}

	return key
}

func stripeWebhookSecretForTest(t *testing.T) string {
	t.Helper()

	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if secret == "" {
		t.Skip("STRIPE_WEBHOOK_SECRET is empty")
	}

	return secret
}
