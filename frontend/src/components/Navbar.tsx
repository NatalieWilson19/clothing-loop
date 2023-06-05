import { Link, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useContext, useState } from "react";

// Project resources
import LanguageSwitcher from "./LanguageSwitcher";
import { AuthContext } from "../providers/AuthProvider";

function Navbar() {
  const { t } = useTranslation();
  const { authUser } = useContext(AuthContext);
  let location = useLocation();

  function onClickMobileNavLink() {
    const el = document.getElementById("header-hamburger") as HTMLInputElement;
    console.log(el.value);

    el.checked = false;
  }

  return (
    <div className="container mx-auto z-50 bg-white flex flex-row justify-between lg:justify-start items-center md:px-20 shadow-none">
      <Link
        aria-label="Clothing Loop logo"
        to="/"
        onClick={onClickMobileNavLink}
        className="bg-center w-32 md:w-40 h-20 md:h-28 bg-no-repeat relative z-[60] bg-[auto_120px] md:bg-[auto_139px]"
        style={{
          backgroundImage:
            "url('https://images.clothingloop.org/x139/the_clothing_loop_logo.png')",
        }}
      ></Link>
      <input type="checkbox" className="hidden peer" id="header-hamburger" />
      <label
        htmlFor="header-hamburger"
        className="mr-3 btn-lg btn-circle btn-ghost hover:bg-base-200 peer-checked:text-secondary peer-checked:animate-[spin-quarter_150ms_linear] flex justify-center items-center lg:hidden relative z-[60] checked:ring-2 checked:ring-offset-2 ring-teal peer-[:checked_>:nth-of-type(1)]:hidden peer-[:checked_>:nth-of-type(2)]:block cursor-pointer"
        aria-label="Menu"
      >
        <span className="feather feather-menu text-2xl block"></span>
        <span className="feather feather-x text-2xl hidden"></span>
      </label>
      <div className="hidden peer-checked:block fixed inset-0 lg:!hidden z-50 bg-white"></div>
      <div className="hidden peer-checked:block absolute inset-0 lg:!hidden z-50">
        <nav
          aria-label="mobile site navigation"
          className="container mx-auto h-screen flex flex-col items-center justify-center"
        >
          <Link
            onClick={onClickMobileNavLink}
            to="/"
            className="mb-3 btn btn-ghost text-base"
          >
            {t("home")}
          </Link>
          <Link
            onClick={onClickMobileNavLink}
            to="/loops/new/users/signup"
            className="mb-3 btn btn-primary btn-outline"
          >
            <span className="feather feather-arrow-left mr-3"></span>
            {t("startNewLoop")}
          </Link>
          <Link
            onClick={onClickMobileNavLink}
            to="/loops/find"
            className="mb-3 btn btn-primary btn-outline"
          >
            {t("findLoops")}
            <span className="feather feather-arrow-right ml-3 rtl:hidden"></span>
            <span className="feather feather-arrow-left mr-3 ltr:hidden"></span>
          </Link>

          <Link
            onClick={onClickMobileNavLink}
            to="/donate"
            className="mb-3 btn btn-ghost text-base"
          >
            {t("donate")}
          </Link>

          <Link
            onClick={onClickMobileNavLink}
            to="/events"
            className="mb-3 btn btn-ghost text-base"
          >
            {t("events")}
          </Link>

          {authUser ? (
            <Link
              onClick={onClickMobileNavLink}
              to="/admin/dashboard"
              className="mb-3 btn btn-ghost text-base"
            >
              {t("account")}
            </Link>
          ) : (
            <Link
              onClick={onClickMobileNavLink}
              to="/users/login"
              className="mb-3 btn btn-ghost text-base"
            >
              {t("login")}
            </Link>
          )}

          <Link
            onClick={onClickMobileNavLink}
            to="/about"
            className="mb-3 btn btn-ghost text-base"
          >
            {t("about")}
          </Link>
          <LanguageSwitcher />
        </nav>
      </div>
      <div className="hidden lg:flex items-center justify-end flex-grow">
        <nav
          aria-label="site navigation"
          className="flex items-center min-h-[4rem]"
        >
          {["/loops/find", "/"].indexOf(location.pathname.substring(3)) !==
          -1 ? (
            <Link
              to="/loops/new/users/signup"
              className="ltr:mr-4 rtl:ml-4 btn btn-primary btn-outline"
            >
              {t("startNewLoop")}
            </Link>
          ) : (
            <Link
              to="/loops/find"
              className="ltr:mr-4 rtl:ml-4 btn btn-primary btn-outline"
            >
              {t("findLoops")}
              <span className="feather feather-arrow-right ml-4 rtl:hidden"></span>
              <span className="feather feather-arrow-left mr-4 ltr:hidden"></span>
            </Link>
          )}

          <Link to="/events" className="btn btn-ghost text-base">
            {t("events")}
          </Link>

          <Link to="/donate" className="btn btn-ghost text-base">
            {t("donate")}
          </Link>

          {authUser ? (
            <Link to="/admin/dashboard" className="btn btn-ghost text-base">
              {t("account")}
            </Link>
          ) : (
            <Link to="/users/login" className="btn btn-ghost text-base">
              {t("login")}
            </Link>
          )}

          <Link to="/about" className="btn btn-ghost text-base">
            {t("about")}
          </Link>

          <LanguageSwitcher className="ltr:ml-4 rtl:mr-4" />
        </nav>
      </div>
    </div>
  );
}

export default Navbar;
