package i18n

// Locale holds all translatable strings for the bot UI and log messages.
type Locale struct {
	// Tabs
	TabOverview    string `json:"tab_overview"`
	TabTrading     string `json:"tab_trading"`
	TabOrders      string `json:"tab_orders"`
	TabPositions   string `json:"tab_positions"`
	TabWallets     string `json:"tab_wallets"`
	TabCopytrading string `json:"tab_copytrading"`
	TabMarkets     string `json:"tab_markets"`
	TabLogs        string `json:"tab_logs"`
	TabSettings    string `json:"tab_settings"`

	// App header
	AppRunning string `json:"app_running"`
	AppWallet  string `json:"app_wallet"`
	HelpGlobal string `json:"help_global"`

	// Overview
	OverviewSubsystems    string `json:"overview_subsystems"`
	OverviewStats         string `json:"overview_stats"`
	OverviewActive        string `json:"overview_active"`
	OverviewInactive      string `json:"overview_inactive"`
	OverviewBalance       string `json:"overview_balance"`
	OverviewOpenOrders    string `json:"overview_open_orders"`
	OverviewPositions     string `json:"overview_positions"`
	OverviewPnLToday      string `json:"overview_pnl_today"`
	OverviewCopyTraders   string `json:"overview_copy_traders"`
	OverviewWallets       string `json:"overview_wallets"`
	OverviewTotalBalance  string `json:"overview_total_balance"`
	OverviewTotalPnL      string `json:"overview_total_pnl"`
	OverviewActiveWallets string `json:"overview_active_wallets"`

	// Health block in Overview
	OverviewHealth        string `json:"overview_health"`
	OverviewHealthUpdated string `json:"overview_health_updated"` // "Updated %ds ago"
	OverviewHealthNever   string `json:"overview_health_never"`
	OverviewGeoBlocked    string `json:"overview_geo_blocked"`
	OverviewGeoAllowed    string `json:"overview_geo_allowed"`

	// Orders tab
	OrdersColMarket string `json:"orders_col_market"`
	OrdersColSide   string `json:"orders_col_side"`
	OrdersColPrice  string `json:"orders_col_price"`
	OrdersColSize   string `json:"orders_col_size"`
	OrdersColFilled string `json:"orders_col_filled"`
	OrdersColStatus string `json:"orders_col_status"`
	OrdersColAge    string `json:"orders_col_age"`
	OrdersHelp      string `json:"orders_help"`
	OrdersEmpty     string `json:"orders_empty"`

	// Positions tab
	PosColMarket  string `json:"pos_col_market"`
	PosColSide    string `json:"pos_col_side"`
	PosColSize    string `json:"pos_col_size"`
	PosColEntry   string `json:"pos_col_entry"`
	PosColCurrent string `json:"pos_col_current"`
	PosColPnL     string `json:"pos_col_pnl"`
	PosColPnLPct  string `json:"pos_col_pnl_pct"`
	PosHelp       string `json:"pos_help"`
	PosEmpty      string `json:"pos_empty"`

	// Copytrading tab
	CopyColAddress   string `json:"copy_col_address"`
	CopyColLabel     string `json:"copy_col_label"`
	CopyColStatus    string `json:"copy_col_status"`
	CopyColAlloc     string `json:"copy_col_alloc"`
	CopyTraders      string `json:"copy_traders"`
	CopyRecentTrades string `json:"copy_recent_trades"`
	CopyNoData       string `json:"copy_no_data"`

	// Logs tab
	LogsFrozen string `json:"logs_frozen"`
	LogsFilter string `json:"logs_filter"`
	LogsHelp   string `json:"logs_help"`

	// Settings sections
	SectionUI            string `json:"section_ui"`
	SectionAuth          string `json:"section_auth"`
	SectionAPI           string `json:"section_api"`
	SectionMonitor       string `json:"section_monitor"`
	SectionTradesMonitor string `json:"section_trades_monitor"`
	SectionTrading       string `json:"section_trading"`
	SectionCopytrading   string `json:"section_copytrading"`
	SectionTelegram      string `json:"section_telegram"`
	SectionDatabase      string `json:"section_database"`
	SectionLog           string `json:"section_log"`
	SectionWebUI         string `json:"section_webui"`

	// Settings field labels
	FieldLanguage         string `json:"field_language"`
	FieldPrivKey          string `json:"field_priv_key"`
	FieldChainID          string `json:"field_chain_id"`
	FieldDefaultOrderType string `json:"field_default_order_type"`
	FieldWebUIListen      string `json:"field_webui_listen"`
	FieldWebUIJWTSecret   string `json:"field_webui_jwt_secret"`
	FieldTimeout        string `json:"field_timeout"`
	FieldMaxRetries     string `json:"field_max_retries"`
	FieldEnabled        string `json:"field_enabled"`
	FieldPollInterval   string `json:"field_poll_interval"`
	FieldAlertOnFill    string `json:"field_alert_on_fill"`
	FieldAlertOnCancel  string `json:"field_alert_on_cancel"`
	FieldTradesLimit    string `json:"field_trades_limit"`
	FieldMaxPositionUSD string `json:"field_max_position_usd"`
	FieldSlippagePct    string `json:"field_slippage_pct"`
	FieldNegRisk        string `json:"field_neg_risk"`
	FieldSizeMode       string `json:"field_size_mode"`
	FieldBotToken       string `json:"field_bot_token"`
	FieldChatID         string `json:"field_chat_id"`
	FieldAdminChatID    string `json:"field_admin_chat_id"`
	FieldDBPath         string `json:"field_db_path"`
	FieldLogLevel       string `json:"field_log_level"`
	FieldLogFormat      string `json:"field_log_format"`
	FieldLogFile        string `json:"field_log_file"`
	FieldTrackPositions string `json:"field_track_positions"`

	// Settings tooltips
	TooltipLanguage        string `json:"tooltip_language"`
	TooltipPrivKey         string `json:"tooltip_priv_key"`
	TooltipChainID         string `json:"tooltip_chain_id"`
	TooltipDefaultOrderType string `json:"tooltip_default_order_type"`
	TooltipWebUIEnabled     string `json:"tooltip_webui_enabled"`
	TooltipWebUIListen      string `json:"tooltip_webui_listen"`
	TooltipWebUIJWTSecret   string `json:"tooltip_webui_jwt_secret"`
	TooltipTimeout         string `json:"tooltip_timeout"`
	TooltipMaxRetries      string `json:"tooltip_max_retries"`
	TooltipMonitorEnabled  string `json:"tooltip_monitor_enabled"`
	TooltipMonitorPoll     string `json:"tooltip_monitor_poll"`
	TooltipTradesEnabled   string `json:"tooltip_trades_enabled"`
	TooltipTradesPoll      string `json:"tooltip_trades_poll"`
	TooltipAlertOnFill     string `json:"tooltip_alert_on_fill"`
	TooltipAlertOnCancel   string `json:"tooltip_alert_on_cancel"`
	TooltipTradesLimit     string `json:"tooltip_trades_limit"`
	TooltipTradesTrack     string `json:"tooltip_trades_track"`
	TooltipTradingEnabled  string `json:"tooltip_trading_enabled"`
	TooltipMaxPosition     string `json:"tooltip_max_position"`
	TooltipSlippage        string `json:"tooltip_slippage"`
	TooltipNegRisk         string `json:"tooltip_neg_risk"`
	TooltipCopyEnabled     string `json:"tooltip_copy_enabled"`
	TooltipCopyPoll        string `json:"tooltip_copy_poll"`
	TooltipSizeMode        string `json:"tooltip_size_mode"`
	TooltipTelegramEnabled string `json:"tooltip_telegram_enabled"`
	TooltipBotToken        string `json:"tooltip_bot_token"`
	TooltipChatID          string `json:"tooltip_chat_id"`
	TooltipAdminChatID     string `json:"tooltip_admin_chat_id"`
	TooltipDBEnabled       string `json:"tooltip_db_enabled"`
	TooltipDBPath          string `json:"tooltip_db_path"`
	TooltipLogLevel        string `json:"tooltip_log_level"`
	TooltipLogFormat       string `json:"tooltip_log_format"`
	TooltipLogFile         string `json:"tooltip_log_file"`

	// Settings UI strings
	SettingsOptions  string `json:"settings_options"`
	SettingsValue    string `json:"settings_value"`
	SettingsUnsaved  string `json:"settings_unsaved"`
	SettingsErrField string `json:"settings_err_field"` // "Error in field «%s»: %v"
	SettingsErrSave  string `json:"settings_err_save"`  // "Save error: %v"

	// Settings help bar
	HelpField      string `json:"help_field"`
	HelpSection    string `json:"help_section"`
	HelpToggle     string `json:"help_toggle"`
	HelpNextOption string `json:"help_next_option"`
	HelpEdit       string `json:"help_edit"`
	HelpSave       string `json:"help_save"`
	HelpReset      string `json:"help_reset"`

	// Wizard
	WizardTitle      string `json:"wizard_title"`
	WizardProgress   string `json:"wizard_progress"`   // "Step %d/%d: %s"
	WizardEmptyField string `json:"wizard_empty_field"`
	WizardWriteError string `json:"wizard_write_error"` // "Config write error: %v"
	WizardContinue   string `json:"wizard_continue"`

	// Wizard step labels
	WizardStep1Label string `json:"wizard_step1_label"`

	// Wizard step hints
	WizardStep1Hint string `json:"wizard_step1_hint"`

	// Log messages (zerolog)
	LogBotStarting          string `json:"log_bot_starting"`
	LogL1Initialized        string `json:"log_l1_initialized"`
	LogTelegramEnabled      string `json:"log_telegram_enabled"`
	LogDatabaseOpened       string `json:"log_database_opened"`
	LogTradesMonitorSkip    string `json:"log_trades_monitor_skip"`
	LogTradesMonitorEnabled string `json:"log_trades_monitor_enabled"`
	LogCopytradingSkipL2    string `json:"log_copytrading_skip_l2"`
	LogCopytradingSkipDB    string `json:"log_copytrading_skip_db"`
	LogCopytradingEnabled   string `json:"log_copytrading_enabled"`
	LogShutdownSignal       string `json:"log_shutdown_signal"`
	LogShuttingDown         string `json:"log_shutting_down"`
	LogBotRunning           string `json:"log_bot_running"`
	LogFatalError           string `json:"log_fatal_error"`
	LogBye                  string `json:"log_bye"`
	LogWSUserEvent          string `json:"log_ws_user_event"`
	LogMonitorStarted       string `json:"log_monitor_started"`
	LogTradesMonitorStarted string `json:"log_trades_monitor_started"`
	LogFailedFetchOrders    string `json:"log_failed_fetch_orders"`
	LogOrdersUpdated        string `json:"log_orders_updated"`
	LogOrderClosed          string `json:"log_order_closed"`
	LogFailedSendAlert      string `json:"log_failed_send_alert"`
	LogNewOrderDetected     string `json:"log_new_order_detected"`
	LogFailedFetchTrades    string `json:"log_failed_fetch_trades"`
	LogTradeExecuted        string `json:"log_trade_executed"`

	// Telegram alert messages
	TgOrderClosed   string `json:"tg_order_closed"`   // "🔔 Order closed: %s"
	TgNewOrder      string `json:"tg_new_order"`      // "📋 New order: %s %s @ %s (size: %s)"
	TgTradeExecuted string `json:"tg_trade_executed"` // "✅ Trade executed: %s %s @ %s (size: %s)"

	// Telegram Bot UI — main menu
	TgWelcome      string `json:"tg_welcome"`       // welcome text body
	TgChooseSection string `json:"tg_choose_section"` // "Choose a section:"
	TgMenuOverview  string `json:"tg_menu_overview"`
	TgMenuTrading   string `json:"tg_menu_trading"`
	TgMenuCopy      string `json:"tg_menu_copy"`
	TgMenuWallets   string `json:"tg_menu_wallets"`
	TgMenuMarkets   string `json:"tg_menu_markets"`
	TgMenuLogs      string `json:"tg_menu_logs"`
	TgMenuSettings  string `json:"tg_menu_settings"`

	// Telegram Bot UI — renderer content
	TgOverviewTitle     string `json:"tg_overview_title"`     // "📊 Overview"
	TgOverviewBalance   string `json:"tg_overview_balance"`   // "💰 Balance: %.2f USDC"
	TgOverviewStats     string `json:"tg_overview_stats"`     // "📋 Orders: %d  |  💼 Positions: %d"
	TgOverviewSubsystems string `json:"tg_overview_subsystems"` // "Subsystems:"
	TgStatusActive      string `json:"tg_status_active"`
	TgStatusInactive    string `json:"tg_status_inactive"`
	TgStatusEnabled     string `json:"tg_status_enabled"`
	TgStatusDisabled    string `json:"tg_status_disabled"`
	TgStatusOn          string `json:"tg_status_on"`  // "🟢 ON"
	TgStatusOff         string `json:"tg_status_off"` // "🔴 OFF"

	TgOrdersTitle    string `json:"tg_orders_title"`    // "📋 Orders (%d)"
	TgOrdersEmpty    string `json:"tg_orders_empty"`    // "📋 Orders\n\nNo open orders."
	TgPositionsTitle string `json:"tg_positions_title"` // "💼 Positions (%d)"
	TgPositionsEmpty string `json:"tg_positions_empty"` // "💼 Positions\n\nNo open positions."

	TgCopyTitle       string `json:"tg_copy_title"`        // "🔄 Copytrading (%d traders)"
	TgCopyEmpty       string `json:"tg_copy_empty"`        // "🔄 Copytrading\n\nNo traders configured."
	TgCopyRecentTrades string `json:"tg_copy_recent_trades"` // "Recent Trades:"

	TgLogsTitle string `json:"tg_logs_title"` // "📝 Logs (last %d)"
	TgLogsEmpty string `json:"tg_logs_empty"` // "📝 Logs\n\nNo log entries yet."

	TgNotSet string `json:"tg_not_set"` // "not set"

	TgMarketsTitle      string `json:"tg_markets_title"`       // "🏪 Markets"
	TgMarketsNA         string `json:"tg_markets_na"`          // "🏪 Markets\n\n<i>Markets service not running.</i>"
	TgMarketsEmpty      string `json:"tg_markets_empty"`       // "<i>No markets available.</i>"
	TgMarketsFilterHint string `json:"tg_markets_filter_hint"` // hint when tag filter active but no results
	TgMarketsShowing    string `json:"tg_markets_showing"`     // "Showing %d of %d markets."
	TgMarketsTapHint    string `json:"tg_markets_tap_hint"`    // "(Tap button below to select market)"
	TgMarketDetail      string `json:"tg_market_detail"`       // "🏪 Market Detail"
	TgMarketLiquidity   string `json:"tg_market_liquidity"`    // "💧 Liquidity: $%.0f"
	TgMarketVolume      string `json:"tg_market_volume"`       // "📊 Volume: $%.0f"
	TgMarketEnds        string `json:"tg_market_ends"`         // "📅 Ends: %s"
	TgMarketCategory    string `json:"tg_market_category"`     // "🏷 Category: %s"

	TgWalletsTitle string `json:"tg_wallets_title"` // "<b>👛 Wallets</b>"
	TgWalletsEmpty string `json:"tg_wallets_empty"` // "<b>👛 Wallets</b>\n\nNo wallets configured."
	TgWalletTotal  string `json:"tg_wallet_total"`  // "Total: $%.2f  P&L: %s%.2f  Active: %d/%d"

	// Telegram Bot UI — keyboard buttons
	TgBtnMainMenu     string `json:"tg_btn_main_menu"`
	TgBtnBackSettings string `json:"tg_btn_back_settings"`
	TgBtnBackMarkets  string `json:"tg_btn_back_market_list"`
	TgBtnBackMarket   string `json:"tg_btn_back_market"`
	TgBtnEnable       string `json:"tg_btn_enable"`
	TgBtnDisable      string `json:"tg_btn_disable"`
	TgBtnRemove       string `json:"tg_btn_remove"`
	TgBtnEdit         string `json:"tg_btn_edit"`
	TgBtnAddWallet    string `json:"tg_btn_add_wallet"`
	TgBtnAddTrader    string `json:"tg_btn_add_trader"`
	TgBtnRefresh      string `json:"tg_btn_refresh"`
	TgBtnConfirm      string `json:"tg_btn_confirm"`
	TgBtnCancel       string `json:"tg_btn_cancel"`
	TgBtnCancelAll    string `json:"tg_btn_cancel_all"`      // "❌ Cancel ALL"
	TgBtnYesCancelAll string `json:"tg_btn_yes_cancel_all"`  // "✅ Yes, cancel all"
	TgBtnNoGoBack     string `json:"tg_btn_no_go_back"`      // "🚫 No, go back"
	TgBtnAllMarkets   string `json:"tg_btn_all_markets"`     // "🌐 All markets"
	TgBtnSetAlert     string `json:"tg_btn_set_alert"`       // "🔔 Set Alert"
	TgBtnFullOrder    string `json:"tg_btn_full_order"`      // "📊 Full Order"
	TgBtnAboveAlert   string `json:"tg_btn_above_alert"`
	TgBtnBelowAlert   string `json:"tg_btn_below_alert"`
	TgBtnBuySide      string `json:"tg_btn_buy_side"`  // "📈 YES (Buy)"
	TgBtnSellSide     string `json:"tg_btn_sell_side"` // "📉 NO (Sell)"
	TgBtnGTC          string `json:"tg_btn_gtc"`
	TgBtnFOK          string `json:"tg_btn_fok"`
	TgBtnOrders       string `json:"tg_btn_orders"`    // "📋 Orders"
	TgBtnPositions    string `json:"tg_btn_positions"` // "💼 Positions"
	TgBtnCancelOrder  string `json:"tg_btn_cancel_order"` // "❌ Cancel #%d (%s)"

	// Telegram Bot UI — screen titles / prompts
	TgTitleSettings      string `json:"tg_title_settings"`
	TgTitleLanguage      string `json:"tg_title_language"`
	TgTitleSetAlert      string `json:"tg_title_set_alert"`
	TgTitleOrderType     string `json:"tg_title_order_type"`
	TgTitleSelectWallet  string `json:"tg_title_select_wallet"`
	TgTitlePlaceOrder    string `json:"tg_title_place_order"`
	TgTitleConfirmOrder  string `json:"tg_title_confirm_order"`  // "📊 Confirm Order\n\nSide: %s\nPrice: %s\nSize: $%s\nType: %s\nWallet: %s"
	TgTitleConfirmQB     string `json:"tg_title_confirm_qb"`     // "📊 Confirm Quick Buy %s\n\nPrice: %.4f\nSize: $%.2f\nWallet: %s\nCost: $%.2f"
	TgTitleCancelConfirm string `json:"tg_title_cancel_confirm"` // "⚠️ Cancel ALL open orders?"

	// Telegram Bot UI — input prompts
	TgInputTraderAddr      string `json:"tg_input_trader_addr"`
	TgInputTraderLabel     string `json:"tg_input_trader_label"`
	TgInputTraderAlloc     string `json:"tg_input_trader_alloc"`
	TgInputEditTraderAlloc string `json:"tg_input_edit_trader_alloc"`
	TgInputMaxPos          string `json:"tg_input_max_pos"`
	TgInputEditKey         string `json:"tg_input_edit_key"`      // "✏️ Enter new value for <code>%s</code>:\n<i>(or /menu to cancel)</i>"
	TgInputPrivKey         string `json:"tg_input_priv_key"`
	TgInputDeleteWallet    string `json:"tg_input_delete_wallet"` // "⚠️ Delete wallet <code>%s</code>? Type <b>yes</b> to confirm:"
	TgInputEditTrader      string `json:"tg_input_edit_trader"`   // "✏️ <b>Edit Trader</b> <code>%s</code>\n\nEnter new label (or <code>-</code> to leave empty):"
	TgInputOrderPrice      string `json:"tg_input_order_price"`   // "📊 Side: <b>%s</b>\n\nEnter price (0.01–0.99):\n<i>(or /menu to cancel)</i>"
	TgInputOrderSize       string `json:"tg_input_order_size"`    // "📊 Price: <b>%.4f</b>\n\nEnter position size in USD:"
	TgInputAlertAbove      string `json:"tg_input_alert_above"`
	TgInputAlertBelow      string `json:"tg_input_alert_below"`
	TgInputQuickBuySize    string `json:"tg_input_quickbuy_size"` // "💚 <b>Quick Buy %s</b>\n<i>%s</i>\n\nPrice: <b>%.4f</b>\n\nEnter bet size in USD:"

	// Telegram Bot UI — error messages
	TgErrAddrEmpty       string `json:"tg_err_addr_empty"`
	TgErrPrivKeyEmpty    string `json:"tg_err_priv_key_empty"`
	TgErrCancelled       string `json:"tg_err_cancelled"`
	TgErrPriceRange      string `json:"tg_err_price_range"`
	TgErrPositiveNum     string `json:"tg_err_positive_num"`
	TgErrNoWallets       string `json:"tg_err_no_wallets"`
	TgErrOrderDataLost   string `json:"tg_err_order_data_lost"`
	TgErrMarketsUnavail  string `json:"tg_err_markets_unavail"`
	TgErrUnknownCmd      string `json:"tg_err_unknown_cmd"`
	TgErrMarketNotFound  string `json:"tg_err_market_not_found"`
	TgErrMarketCtxLost   string `json:"tg_err_market_ctx_lost"`
	TgErrAlertDataFmt    string `json:"tg_err_alert_data_fmt"`
	TgErrOrderUnavail    string `json:"tg_err_order_unavail"`
	TgErrOrderCorrupt    string `json:"tg_err_order_corrupt"`
	TgErrOrderPlace      string `json:"tg_err_order_place"` // "Order placement error: %s"
	TgErrTraderExists    string `json:"tg_err_trader_exists"`   // "Trader %q already exists."
	TgErrTraderNotFound  string `json:"tg_err_trader_not_found"` // "Trader %q not found."
	TgErrSaveFailed      string `json:"tg_err_save_failed"`     // "Failed to save config: %v"
	TgErrKeyAdmin        string `json:"tg_err_key_admin"`       // "Key %q requires admin access."
	TgErrKeyInvalid      string `json:"tg_err_key_invalid"`     // "Invalid value for %q: %v"
	TgErrKeyUnknown      string `json:"tg_err_key_unknown"`     // "Unknown key: %q"
	TgErrWalletExists    string `json:"tg_err_wallet_exists"`   // "Wallet already exists: %s"
	TgErrInvalidPrivKey  string `json:"tg_err_invalid_priv_key"` // "Invalid private key: %s"
	TgErrWalletManagerNA string `json:"tg_err_wallet_manager_na"`
	TgErrRemoveNotSupported string `json:"tg_err_remove_not_supported"`
	TgErrUnknownSection  string `json:"tg_err_unknown_section"` // "Unknown section: %s"
	TgErrCancelUnavail   string `json:"tg_err_cancel_unavail"`

	// Telegram Bot UI — success messages
	TgSuccessOrderPlaced   string `json:"tg_success_order_placed"`   // "Order placed!\n\nID: <code>%s</code>\nSide: <b>%s</b> | Price: <b>%.4f</b> | Size: <b>$%.2f</b>"
	TgSuccessAlertCreated  string `json:"tg_success_alert_created"`  // "Alert created! %s Price %s <b>%.3f</b>\n<code>ID: %s</code>"
	TgSuccessTraderAdded   string `json:"tg_success_trader_added"`   // "Trader <code>%s</code> added (label: %s, alloc: %.1f%%)."
	TgSuccessTraderRemoved string `json:"tg_success_trader_removed"` // "Trader <code>%s</code> removed."
	TgSuccessTraderUpdated string `json:"tg_success_trader_updated"` // "Trader <code>%s</code> updated.\nlabel: %s | alloc: %.1f%% | max: $%.0f"
	TgSuccessTraderToggled string `json:"tg_success_trader_toggled"` // "Trader <code>%s</code> %s."
	TgSuccessWalletAdded   string `json:"tg_success_wallet_added"`   // "Wallet added.\nAddress: <code>%s</code>\nID: <code>%s</code>"
	TgSuccessWalletRemoved string `json:"tg_success_wallet_removed"` // "Wallet <code>%s</code> removed."
	TgSuccessWalletToggled string `json:"tg_success_wallet_toggled"` // "Wallet <code>%s</code> %s."
	TgSuccessOrderCancelled string `json:"tg_success_order_cancelled"` // "Order <code>%s</code> cancelled."
	TgSuccessAllCancelled  string `json:"tg_success_all_cancelled"`
	TgSuccessConfigSaved   string `json:"tg_success_config_saved"`   // "<code>%s</code> = <code>%s</code>  Config saved."

	// Telegram health section in Overview
	TgHealthTitle      string `json:"tg_health_title"`
	TgHealthUpdated    string `json:"tg_health_updated"`    // "Updated %ds ago"
	TgHealthNever      string `json:"tg_health_never"`
	TgBtnHealthRefresh string `json:"tg_btn_health_refresh"`
}
