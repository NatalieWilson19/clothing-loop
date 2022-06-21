import "./App.css";
import React from "react";
import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import {
  createTheme,
  ThemeProvider as MuiThemeProvider,
  StyledEngineProvider,
} from "@mui/material/styles";
import { AuthProvider } from "./components/AuthProvider";
import themeFile from "./util/theme";
import ScrollToTop from "./util/scrollToTop";

// Components
import Navbar from "./components/Navbar";
import Footer from "./components/Footer";
import { Logout } from "./pages/Logout";
import { ChainsProvider } from "./components/ChainsProvider";

// Pages
import {
  NewLoopConfirmation,
  JoinLoopConfirmation,
} from "./pages/Thankyou/Thankyou";
const FindChain = React.lazy(() => import("./pages/FindChain"));
const Login = React.lazy(() => import("./pages/Login"));
const ChainMemberList = React.lazy(() => import("./pages/ChainMemberList"));
const NewChainSignup = React.lazy(() => import("./pages/NewChainSignup"));
const Signup = React.lazy(() => import("./pages/Signup"));
const NewChainLocation = React.lazy(() => import("./pages/NewChainLocation"));
const UserEdit = React.lazy(() => import("./pages/UserEdit"));
const ChainEdit = React.lazy(() => import("./pages/ChainEdit"));
const ChainsList = React.lazy(() => import("./pages/ChainsList"));
const Home = React.lazy(() => import("./pages/Home"));
const LoginEmailFinished = React.lazy(
  () => import("./pages/LoginEmailFinished")
);
const Contacts = React.lazy(() => import("./pages/Contacts"));
const MessageSubmitted = React.lazy(() => import("./pages/MessageSubmitted"));
const Donate = React.lazy(() => import("./pages/Donations/Donate"));
const About = React.lazy(() => import("./pages/About"));
const PrivacyPolicy = React.lazy(() => import("./pages/PrivacyPolicy"));
const TermsOfUse = React.lazy(() => import("./pages/TermsOfUse"));
const FAQ = React.lazy(() => import("./pages/FAQ/FAQ"));
const AdminControlsNav = React.lazy(
  () => import("./components/AdminControlsNav/AdminControlsNav")
);
const AddChainAdmin = React.lazy(() => import("./pages/AddChainAdmin"));

const theme = createTheme(themeFile);

const App = () => {
  return (
    <StyledEngineProvider injectFirst>
      <MuiThemeProvider theme={theme}>
        <AuthProvider>
          <ChainsProvider>
            <div className="tw-min-h-screen">
              <Router>
                <ScrollToTop>
                  <Navbar />
                  <Switch>
                    <Route exact path="/" component={Home} />
                    <Route
                      exact
                      path="/thankyou"
                      component={JoinLoopConfirmation}
                    />
                    <Route exact path="/donate/:status?" component={Donate} />
                    <Route
                      exact
                      path="/message-submitted"
                      component={MessageSubmitted}
                    />

                    <Route
                      exact
                      path="/users/login-email-finished/:email"
                      component={LoginEmailFinished}
                    />
                    <Route exact path="/users/login" component={Login} />
                    <Route exact path="/users/logout" component={Logout} />
                    <Route
                      exact
                      path="/users/:userId/edit"
                      component={UserEdit}
                    />

                    <Route exact path="/loops" component={ChainsList} />
                    <Route exact path="/loops/find" component={FindChain} />
                    <Route
                      exact
                      path="/loops/:chainId/edit"
                      component={ChainEdit}
                    />
                    <Route
                      exact
                      path="/loops/:chainId/members"
                      component={ChainMemberList}
                    />
                    <Route
                      exact
                      path="/loops/:chainId/addChainAdmin"
                      component={AddChainAdmin}
                    />
                    <Route
                      exact
                      path="/loops/new/users/signup"
                      component={NewChainSignup}
                    />
                    <Route
                      exact
                      path="/loops/new"
                      component={NewChainLocation}
                    />
                    <Route
                      exact
                      path="/loops/new/confirmation"
                      component={NewLoopConfirmation}
                    />
                    <Route
                      exact
                      path="/loops/:chainId/users/signup"
                      component={Signup}
                    />

                    <Route exact path="/faq" component={FAQ} />
                    <Route exact path="/contact-us" component={Contacts} />
                    <Route exact path="/about" component={About} />

                    <Route exact path="/terms-of-use" component={TermsOfUse} />
                    <Route
                      exact
                      path="/privacy-policy"
                      component={PrivacyPolicy}
                    />

                    <Route
                      exact
                      path="/admin/dashboard"
                      component={AdminControlsNav}
                    />
                  </Switch>
                  <Footer />
                </ScrollToTop>
              </Router>
            </div>
          </ChainsProvider>
        </AuthProvider>
      </MuiThemeProvider>
    </StyledEngineProvider>
  );
};

export default App;
