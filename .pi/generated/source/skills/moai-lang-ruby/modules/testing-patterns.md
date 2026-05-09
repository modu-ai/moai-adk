# Ruby Testing Patterns

## RSpec Model Specs

Complete Model Testing:
```ruby
# spec/models/user_spec.rb
RSpec.describe User, type: :model do
  describe "associations" do
    it { is_expected.to have_many(:posts).dependent(:destroy) }
    it { is_expected.to have_one(:profile).dependent(:destroy) }
  end

  describe "validations" do
    subject { build(:user) }

    it { is_expected.to validate_presence_of(:email) }
    it { is_expected.to validate_uniqueness_of(:email).case_insensitive }
    it { is_expected.to validate_length_of(:name).is_at_least(2).is_at_most(100) }
  end

  describe "scopes" do
    describe ".active" do
      let!(:active_user) { create(:user, active: true) }
      let!(:inactive_user) { create(:user, active: false) }

      it "returns only active users" do
        expect(described_class.active).to contain_exactly(active_user)
      end
    end
  end

  describe "#full_name" do
    context "when both names are present" do
      let(:user) { build(:user, first_name: "John", last_name: "Doe") }

      it "returns the full name" do
        expect(user.full_name).to eq("John Doe")
      end
    end

    context "when last name is missing" do
      let(:user) { build(:user, first_name: "John", last_name: nil) }

      it "returns only first name" do
        expect(user.full_name).to eq("John")
      end
    end
  end
end
```

## RSpec Request Specs

API Testing:
```ruby
# spec/requests/posts_spec.rb
RSpec.describe "Posts", type: :request do
  let(:user) { create(:user) }

  before { sign_in user }

  describe "GET /posts" do
    let!(:posts) { create_list(:post, 3, user: user) }

    it "returns a successful response" do
      get posts_path
      expect(response).to have_http_status(:ok)
    end

    it "displays all posts" do
      get posts_path
      posts.each do |post|
        expect(response.body).to include(post.title)
      end
    end
  end

  describe "POST /posts" do
    let(:valid_params) { { post: attributes_for(:post) } }
    let(:invalid_params) { { post: { title: "" } } }

    context "with valid parameters" do
      it "creates a new post" do
        expect {
          post posts_path, params: valid_params
        }.to change(Post, :count).by(1)
      end

      it "redirects to the created post" do
        post posts_path, params: valid_params
        expect(response).to redirect_to(Post.last)
      end
    end

    context "with invalid parameters" do
      it "does not create a new post" do
        expect {
          post posts_path, params: invalid_params
        }.not_to change(Post, :count)
      end

      it "returns unprocessable entity status" do
        post posts_path, params: invalid_params
        expect(response).to have_http_status(:unprocessable_entity)
      end
    end
  end
end
```

## Factory Bot Patterns

Advanced Factories:
```ruby
# spec/factories/users.rb
FactoryBot.define do
  factory :user do
    sequence(:email) { |n| "user#{n}@example.com" }
    name { Faker::Name.name }
    password { "password123" }
    active { true }

    trait :inactive do
      active { false }
    end

    trait :admin do
      role { :admin }
    end

    trait :with_posts do
      transient do
        posts_count { 3 }
      end

      after(:create) do |user, evaluator|
        create_list(:post, evaluator.posts_count, user: user)
      end
    end

    factory :admin_user, traits: [:admin]
  end
end

# spec/factories/posts.rb
FactoryBot.define do
  factory :post do
    sequence(:title) { |n| "Post Title #{n}" }
    content { Faker::Lorem.paragraphs(number: 3).join("\n\n") }
    user
    published { false }

    trait :published do
      published { true }
      published_at { Time.current }
    end

    trait :with_comments do
      transient do
        comments_count { 5 }
      end

      after(:create) do |post, evaluator|
        create_list(:comment, evaluator.comments_count, post: post)
      end
    end
  end
end
```

## System Specs with Capybara

End-to-End Testing:
```ruby
# spec/system/user_registration_spec.rb
RSpec.describe "User Registration", type: :system do
  before do
    driven_by(:selenium_chrome_headless)
  end

  it "allows a user to register" do
    visit new_user_registration_path

    fill_in "Email", with: "newuser@example.com"
    fill_in "Password", with: "password123"
    fill_in "Password confirmation", with: "password123"
    click_button "Sign up"

    expect(page).to have_content("Welcome! You have signed up successfully.")
    expect(User.find_by(email: "newuser@example.com")).to be_present
  end

  it "shows validation errors for invalid input" do
    visit new_user_registration_path

    fill_in "Email", with: "invalid-email"
    click_button "Sign up"

    expect(page).to have_content("Email is invalid")
  end
end
```

## Service Object Testing

Testing Service Layer:
```ruby
# spec/services/user_registration_service_spec.rb
RSpec.describe UserRegistrationService do
  describe "#call" do
    let(:valid_params) do
      {
        name: "John Doe",
        email: "john@example.com",
        password: "password123"
      }
    end

    context "with valid parameters" do
      it "creates a user" do
        result = described_class.new(valid_params).call

        expect(result.success?).to be true
        expect(result.user).to be_persisted
      end

      it "creates a profile for the user" do
        result = described_class.new(valid_params).call

        expect(result.user.profile).to be_present
      end

      it "sends a welcome email" do
        expect {
          described_class.new(valid_params).call
        }.to have_enqueued_mail(UserMailer, :welcome)
      end
    end

    context "with invalid parameters" do
      let(:invalid_params) { { email: "invalid" } }

      it "returns a failure result" do
        result = described_class.new(invalid_params).call

        expect(result.failure?).to be true
        expect(result.errors).to be_present
      end

      it "does not create a user" do
        expect {
          described_class.new(invalid_params).call
        }.not_to change(User, :count)
      end
    end
  end
end
```

## Shared Examples

Reusable Test Patterns:
```ruby
# spec/support/shared_examples/api_resource.rb
RSpec.shared_examples "an API resource" do
  it "returns JSON content type" do
    expect(response.content_type).to include("application/json")
  end

  it "returns a successful response" do
    expect(response).to have_http_status(:ok)
  end
end

# Usage in specs
RSpec.describe "Users API", type: :request do
  describe "GET /api/users" do
    before { get api_users_path }

    it_behaves_like "an API resource"
  end
end
```

## Mocking and Stubbing

External Service Mocking:
```ruby
RSpec.describe PaymentService do
  describe "#charge" do
    let(:payment_gateway) { instance_double(Stripe::PaymentIntent) }

    before do
      allow(Stripe::PaymentIntent).to receive(:create).and_return(payment_gateway)
      allow(payment_gateway).to receive(:status).and_return("succeeded")
    end

    it "processes the payment successfully" do
      result = described_class.new.charge(amount: 1000, customer_id: "cus_123")

      expect(result).to be_success
    end
  end
end
```
