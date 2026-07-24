begin
  require 'sidekiq'
rescue LoadError
end

module OrcaStack
  module RubyApp
    module Jobs
      class WebhookDeliveryJob
        if defined?(Sidekiq::Job)
          include Sidekiq::Job
          sidekiq_options queue: 'webhooks', retry: 5
        end

        def self.enqueue(webhook_id)
          if respond_to?(:perform_async)
            perform_async(webhook_id)
          else
            new.perform(webhook_id)
          end
        end

        def perform(webhook_id)
          Store.mark_webhook_delivered(webhook_id)
        end
      end
    end
  end
end
