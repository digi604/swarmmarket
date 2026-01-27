-- SwarmMarket Domain Taxonomy Seed
-- Migration: 003_seed_taxonomy

-- ============================================
-- DELIVERY
-- ============================================
INSERT INTO domain_taxonomy (path, parent_path, name, description, icon) VALUES
('delivery', NULL, 'Delivery', 'Physical delivery of goods and items', '📦'),
('delivery/food', 'delivery', 'Food Delivery', 'Restaurant and prepared food delivery', '🍕'),
('delivery/food/restaurant', 'delivery/food', 'Restaurant', 'Delivery from restaurants', '🍽️'),
('delivery/food/grocery', 'delivery/food', 'Grocery', 'Grocery and supermarket delivery', '🛒'),
('delivery/food/catering', 'delivery/food', 'Catering', 'Large order and event catering', '🎉'),
('delivery/packages', 'delivery', 'Packages', 'Package and parcel delivery', '📦'),
('delivery/packages/same_day', 'delivery/packages', 'Same Day', 'Same-day package delivery', '⚡'),
('delivery/packages/next_day', 'delivery/packages', 'Next Day', 'Next-day package delivery', '📅'),
('delivery/packages/international', 'delivery/packages', 'International', 'International shipping', '🌍'),
('delivery/documents', 'delivery', 'Documents', 'Document and letter delivery', '📄');

-- ============================================
-- DATA
-- ============================================
INSERT INTO domain_taxonomy (path, parent_path, name, description, icon) VALUES
('data', NULL, 'Data', 'Data collection, processing, and analysis', '📊'),
('data/web', 'data', 'Web Data', 'Web-based data services', '🌐'),
('data/web/scraping', 'data/web', 'Scraping', 'Web scraping and extraction', '🕷️'),
('data/web/search', 'data/web', 'Search', 'Web search and research', '🔍'),
('data/web/monitoring', 'data/web', 'Monitoring', 'Website and content monitoring', '👁️'),
('data/analysis', 'data', 'Analysis', 'Data analysis services', '📈'),
('data/analysis/sentiment', 'data/analysis', 'Sentiment', 'Sentiment analysis', '😊'),
('data/analysis/summarization', 'data/analysis', 'Summarization', 'Content summarization', '📝'),
('data/analysis/extraction', 'data/analysis', 'Extraction', 'Entity and data extraction', '🎯'),
('data/generation', 'data', 'Generation', 'Content generation', '✨'),
('data/generation/text', 'data/generation', 'Text', 'Text generation', '📝'),
('data/generation/image', 'data/generation', 'Image', 'Image generation', '🖼️'),
('data/generation/code', 'data/generation', 'Code', 'Code generation', '💻');

-- ============================================
-- SERVICES
-- ============================================
INSERT INTO domain_taxonomy (path, parent_path, name, description, icon) VALUES
('services', NULL, 'Services', 'Service-based offerings', '🛎️'),
('services/booking', 'services', 'Booking', 'Reservation and booking services', '📅'),
('services/booking/restaurants', 'services/booking', 'Restaurants', 'Restaurant reservations', '🍽️'),
('services/booking/travel', 'services/booking', 'Travel', 'Travel and hotel booking', '✈️'),
('services/booking/appointments', 'services/booking', 'Appointments', 'Appointment scheduling', '📆'),
('services/communication', 'services', 'Communication', 'Communication services', '💬'),
('services/communication/email', 'services/communication', 'Email', 'Email sending and management', '📧'),
('services/communication/sms', 'services/communication', 'SMS', 'SMS and text messaging', '📱'),
('services/communication/calls', 'services/communication', 'Calls', 'Phone calls and voice', '📞'),
('services/financial', 'services', 'Financial', 'Financial services', '💰'),
('services/financial/payments', 'services/financial', 'Payments', 'Payment processing', '💳'),
('services/financial/invoicing', 'services/financial', 'Invoicing', 'Invoice generation', '🧾'),
('services/financial/accounting', 'services/financial', 'Accounting', 'Accounting services', '📊');

-- ============================================
-- COMPUTE
-- ============================================
INSERT INTO domain_taxonomy (path, parent_path, name, description, icon) VALUES
('compute', NULL, 'Compute', 'Computational resources and processing', '🖥️'),
('compute/inference', 'compute', 'Inference', 'AI model inference', '🤖'),
('compute/inference/llm', 'compute/inference', 'LLM', 'Large language model inference', '💭'),
('compute/inference/vision', 'compute/inference', 'Vision', 'Computer vision inference', '👁️'),
('compute/inference/audio', 'compute/inference', 'Audio', 'Audio processing and STT/TTS', '🔊'),
('compute/training', 'compute', 'Training', 'Model training and fine-tuning', '🎓'),
('compute/processing', 'compute', 'Processing', 'General data processing', '⚙️');

-- ============================================
-- AUTOMATION
-- ============================================
INSERT INTO domain_taxonomy (path, parent_path, name, description, icon) VALUES
('automation', NULL, 'Automation', 'Task automation and workflows', '🤖'),
('automation/browser', 'automation', 'Browser', 'Browser automation', '🌐'),
('automation/browser/navigation', 'automation/browser', 'Navigation', 'Web navigation and interaction', '🖱️'),
('automation/browser/form_filling', 'automation/browser', 'Form Filling', 'Automated form completion', '📝'),
('automation/browser/testing', 'automation/browser', 'Testing', 'Automated testing', '🧪'),
('automation/workflow', 'automation', 'Workflow', 'Workflow automation', '🔄'),
('automation/workflow/scheduling', 'automation/workflow', 'Scheduling', 'Task scheduling', '⏰'),
('automation/workflow/triggers', 'automation/workflow', 'Triggers', 'Event-based triggers', '⚡');

-- ============================================
-- MARKETPLACE
-- ============================================
INSERT INTO domain_taxonomy (path, parent_path, name, description, icon) VALUES
('marketplace', NULL, 'Marketplace', 'Buying, selling, trading', '🏪'),
('marketplace/pricing', 'marketplace', 'Pricing', 'Price comparison and monitoring', '💲'),
('marketplace/trading', 'marketplace', 'Trading', 'Asset trading', '📈'),
('marketplace/auctions', 'marketplace', 'Auctions', 'Auction participation', '🔨');
