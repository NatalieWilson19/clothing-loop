import { useContext, useState } from "react";
import { Helmet } from "react-helmet";
import { Redirect, useHistory } from "react-router-dom";
import { useTranslation } from "react-i18next";

import ProgressBar from "../components/ProgressBar";
import ChainDetailsForm, {
  RegisterChainForm,
} from "../components/ChainDetailsForm";
import {
  registerChainAdmin,
  RequestRegisterChain,
  RequestRegisterUser,
} from "../api/login";
import { AuthContext } from "../providers/AuthProvider";
import { chainCreate } from "../api/chain";
import { ToastContext } from "../providers/ToastProvider";
import { GinParseErrors } from "../util/gin-errors";

export interface State {
  only_create_chain: boolean;
  register_user?: RequestRegisterUser;
}

const NewChainLocation = ({ location }: { location: any }) => {
  const history = useHistory();
  const { t } = useTranslation();
  const state = location.state as State | undefined;
  const { authUser, authUserRefresh } = useContext(AuthContext);
  const { addToastError } = useContext(ToastContext);

  const onSubmit = async (values: RegisterChainForm) => {
    let user = state!.only_create_chain ? authUser : state!.register_user;
    if (!user) {
      addToastError("User is not availible", 400);
      return;
    }
    const newChain: RequestRegisterChain = {
      name: values.name,
      description: values.description,
      address: values.address,
      latitude: values.latitude,
      longitude: values.longitude,
      radius: values.radius,
      open_to_new_members: true,
      sizes: values.sizes,
      genders: values.genders,
    };

    if (!(newChain.address?.length > 5)) {
      addToastError(t("required") + ": " + t("loopLocation"), 400);
      return;
    }

    console.log(`creating chain: ${JSON.stringify(newChain)}`);
    if (state!.only_create_chain) {
      try {
        await chainCreate(newChain);
        await authUserRefresh();

        if (window.goatcounter)
          window.goatcounter.count({
            path: "new-chain",
            title: "New chain",
            event: true,
          });
        history.replace("/loops/new/confirmation");
      } catch (err: any) {
        console.error("Error creating chain:", err, newChain);
        addToastError(GinParseErrors(t, err), err?.status);
      }
    } else {
      console.log(`creating user: ${JSON.stringify(user)}`);
      try {
        await registerChainAdmin(
          {
            name: user.name,
            email: user.email,
            address: user.address,
            phone_number: user.phone_number,
            newsletter: state!.register_user?.newsletter || false,
            sizes: values.sizes || [],
            latitude: user.latitude,
            longitude: user.longitude,
          },
          newChain
        );
        if (window.goatcounter) {
          window.goatcounter.count({
            path: "new-chain",
            title: "New chain",
            event: true,
          });
          window.goatcounter.count({
            path: "new-user",
            title: "New user",
            event: true,
          });
        }
        history.replace("/loops/new/confirmation");
      } catch (err: any) {
        console.error(`Error creating user and chain: ${JSON.stringify(err)}`);
        addToastError(GinParseErrors(t, err), err?.status);
      }
    }
  };

  if (!state || (state.only_create_chain === false && !state.register_user)) {
    return <Redirect to="/loops/new/users/signup" />;
  }

  return (
    <>
      <Helmet>
        <title>The Clothing Loop | Create New Loop</title>
        <meta name="description" content="Create New Loop" />
      </Helmet>
      <main className="container lg:max-w-screen-lg mx-auto md:px-20 pt-4">
        <div className="bg-teal-light p-8">
          <h1 className="text-center font-medium text-secondary text-5xl mb-6">
            {t("startNewLoop")}
          </h1>
          <div className="text-center mb-6">
            <ProgressBar
              activeStep={1}
              disabledStep={state.only_create_chain ? 0 : undefined}
            />
          </div>
          <ChainDetailsForm
            onSubmit={onSubmit}
            //submitError={error}
            showBack={!state.only_create_chain}
          />
        </div>
      </main>
    </>
  );
};

export default NewChainLocation;
