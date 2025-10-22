# moai-lang-haskell - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with Stack

```bash
# Install Stack (recommended via ghcup)
ghcup install stack

# Create new project
stack new my-project
cd my-project

# Install dependencies
stack setup
stack build

# Run tests
stack test

# Add HUnit for testing
stack install HUnit
```

**package.yaml configuration**:
```yaml
name: my-project
version: 0.1.0.0
dependencies:
  - base >= 4.7 && < 5
  - HUnit >= 1.6.2

tests:
  my-project-test:
    main: Spec.hs
    source-dirs: test
    dependencies:
      - my-project
      - HUnit
```

## Example 2: TDD Workflow with HUnit

**RED: Write failing test**
```haskell
-- test/CalculatorSpec.hs
module CalculatorSpec where

import Test.HUnit
import Calculator

testAddPositive :: Test
testAddPositive = TestCase $ assertEqual
  "should add two positive numbers"
  5
  (add 2 3)

testAddNegative :: Test
testAddNegative = TestCase $ assertEqual
  "should handle negative numbers"
  (-3)
  (add (-1) (-2))

testAddZero :: Test
testAddZero = TestCase $ assertEqual
  "should handle zero"
  5
  (add 0 5)

tests :: Test
tests = TestList
  [ TestLabel "testAddPositive" testAddPositive
  , TestLabel "testAddNegative" testAddNegative
  , TestLabel "testAddZero" testAddZero
  ]

main :: IO ()
main = do
  _ <- runTestTT tests
  return ()
```

**GREEN: Implement feature**
```haskell
-- src/Calculator.hs
module Calculator (add) where

add :: Num a => a -> a -> a
add x y = x + y
```

**REFACTOR: Improve with type safety**
```haskell
-- src/Calculator.hs
{-# LANGUAGE ScopedTypeVariables #-}

module Calculator
  ( add
  , SafeNum(..)
  ) where

-- Type-safe arithmetic
newtype SafeNum = SafeNum Integer
  deriving (Show, Eq)

instance Num SafeNum where
  (SafeNum x) + (SafeNum y) = SafeNum (x + y)
  (SafeNum x) * (SafeNum y) = SafeNum (x * y)
  abs (SafeNum x) = SafeNum (abs x)
  signum (SafeNum x) = SafeNum (signum x)
  fromInteger = SafeNum
  negate (SafeNum x) = SafeNum (negate x)

-- | Add two numbers with type safety
-- >>> add (SafeNum 2) (SafeNum 3)
-- SafeNum 5
add :: Num a => a -> a -> a
add x y = x + y
```

## Example 3: Cabal Configuration

**my-project.cabal**:
```cabal
cabal-version: 2.4
name: my-project
version: 0.1.0.0
build-type: Simple

common warnings
  ghc-options: -Wall
              -Wcompat
              -Widentities
              -Wincomplete-record-updates
              -Wincomplete-uni-patterns
              -Wmissing-home-modules
              -Wpartial-fields
              -Wredundant-constraints

library
  import: warnings
  exposed-modules: Calculator
  build-depends: base ^>=4.18.0.0
  hs-source-dirs: src
  default-language: Haskell2010

test-suite my-project-test
  import: warnings
  default-language: Haskell2010
  type: exitcode-stdio-1.0
  hs-source-dirs: test
  main-is: Spec.hs
  build-depends:
      base ^>=4.18.0.0
    , my-project
    , HUnit >= 1.6.2
```

**Run tests with Cabal**:
```bash
# Configure project
cabal configure --enable-tests

# Build project
cabal build

# Run tests
cabal test

# Run tests with coverage
cabal test --enable-coverage
```

## Example 4: Pure Functions and Immutability

```haskell
-- src/UserService.hs
module UserService
  ( User(..)
  , createUser
  , updateEmail
  , validateUser
  ) where

import Data.Text (Text)
import qualified Data.Text as T

-- Immutable data structure
data User = User
  { userId :: Int
  , userName :: Text
  , userEmail :: Text
  } deriving (Show, Eq)

-- Pure function: no side effects
createUser :: Int -> Text -> Text -> Maybe User
createUser uid name email
  | T.null name = Nothing
  | not (isValidEmail email) = Nothing
  | otherwise = Just $ User uid name email

-- Pure function: returns new user
updateEmail :: User -> Text -> Maybe User
updateEmail user newEmail
  | isValidEmail newEmail = Just $ user { userEmail = newEmail }
  | otherwise = Nothing

-- Pure validation
isValidEmail :: Text -> Bool
isValidEmail email = T.elem '@' email && T.length email > 3

-- Validation with Either for error handling
validateUser :: User -> Either Text User
validateUser user
  | T.null (userName user) = Left "Name cannot be empty"
  | not (isValidEmail (userEmail user)) = Left "Invalid email"
  | otherwise = Right user
```

**Test Suite**:
```haskell
-- test/UserServiceSpec.hs
module UserServiceSpec where

import Test.HUnit
import UserService
import Data.Text (pack)

testCreateValidUser :: Test
testCreateValidUser = TestCase $ do
  let result = createUser 1 (pack "John") (pack "john@example.com")
  assertEqual "should create valid user" True (isJust result)

testCreateInvalidEmail :: Test
testCreateInvalidEmail = TestCase $ do
  let result = createUser 1 (pack "John") (pack "invalid")
  assertEqual "should reject invalid email" Nothing result

testUpdateEmail :: Test
testUpdateEmail = TestCase $ do
  let user = User 1 (pack "John") (pack "old@example.com")
  let result = updateEmail user (pack "new@example.com")
  assertEqual "should update email" 
    (Just (pack "new@example.com")) 
    (fmap userEmail result)

tests :: Test
tests = TestList
  [ TestLabel "testCreateValidUser" testCreateValidUser
  , TestLabel "testCreateInvalidEmail" testCreateInvalidEmail
  , TestLabel "testUpdateEmail" testUpdateEmail
  ]
```

## Example 5: Type-Driven Development

```haskell
-- src/ApiClient.hs
{-# LANGUAGE OverloadedStrings #-}

module ApiClient
  ( ApiResponse(..)
  , fetchUser
  , handleResponse
  ) where

import Data.Text (Text)
import Control.Monad.IO.Class (liftIO)
import Control.Exception (try, SomeException)

-- Type-driven design: make invalid states unrepresentable
data ApiResponse a
  = Success a
  | NotFound
  | ServerError Text
  | NetworkError Text
  deriving (Show, Eq)

-- Type-safe API call
fetchUser :: Int -> IO (ApiResponse User)
fetchUser userId = do
  result <- try (makeHttpRequest userId) :: IO (Either SomeException User)
  case result of
    Right user -> return $ Success user
    Left ex -> return $ NetworkError (pack $ show ex)

-- Exhaust ive pattern matching ensures all cases handled
handleResponse :: ApiResponse a -> (a -> IO ()) -> IO ()
handleResponse response onSuccess = case response of
  Success value -> onSuccess value
  NotFound -> putStrLn "User not found"
  ServerError msg -> putStrLn $ "Server error: " ++ show msg
  NetworkError msg -> putStrLn $ "Network error: " ++ show msg

-- Mock HTTP request for testing
makeHttpRequest :: Int -> IO User
makeHttpRequest uid = return $ User uid "John" "john@example.com"
```

## Example 6: Property-Based Testing with QuickCheck

```haskell
-- test/PropertyTests.hs
module PropertyTests where

import Test.QuickCheck
import Calculator

-- Property: addition is commutative
prop_addCommutative :: Int -> Int -> Bool
prop_addCommutative x y = add x y == add y x

-- Property: addition is associative
prop_addAssociative :: Int -> Int -> Int -> Bool
prop_addAssociative x y z = add (add x y) z == add x (add y z)

-- Property: zero is identity
prop_addIdentity :: Int -> Bool
prop_addIdentity x = add x 0 == x && add 0 x == x

-- Run property tests
main :: IO ()
main = do
  putStrLn "Testing commutativity..."
  quickCheck prop_addCommutative
  
  putStrLn "Testing associativity..."
  quickCheck prop_addAssociative
  
  putStrLn "Testing identity..."
  quickCheck prop_addIdentity
```

## Example 7: Monadic Error Handling

```haskell
-- src/Database.hs
{-# LANGUAGE OverloadedStrings #-}

module Database
  ( DbError(..)
  , DbResult
  , insertUser
  , findUser
  ) where

import Data.Text (Text)
import Control.Monad.Except

data DbError
  = ConnectionError Text
  | QueryError Text
  | NotFoundError
  deriving (Show, Eq)

type DbResult a = ExceptT DbError IO a

-- Monadic error handling
insertUser :: User -> DbResult Int
insertUser user = do
  -- Validate before insert
  case validateUser user of
    Left err -> throwError (QueryError err)
    Right _ -> liftIO $ mockInsert user

findUser :: Int -> DbResult User
findUser userId = do
  result <- liftIO $ mockFind userId
  case result of
    Nothing -> throwError NotFoundError
    Just user -> return user

-- Mock database operations
mockInsert :: User -> IO Int
mockInsert _ = return 1

mockFind :: Int -> IO (Maybe User)
mockFind _ = return $ Just (User 1 "John" "john@example.com")
```

**Usage**:
```haskell
-- Main.hs
main :: IO ()
main = do
  result <- runExceptT $ do
    userId <- insertUser (User 0 "John" "john@example.com")
    findUser userId
  
  case result of
    Left err -> putStrLn $ "Error: " ++ show err
    Right user -> putStrLn $ "Success: " ++ show user
```

---

_For complete API reference and configuration options, see reference.md_
