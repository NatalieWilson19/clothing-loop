// React / plugins
import { useState, useEffect, FormEvent, useContext } from "react";
import { useTranslation } from "react-i18next";
import { Redirect, useParams, useHistory } from "react-router-dom";
import { Helmet } from "react-helmet";

import GeocoderSelector from "../components/GeocoderSelector";
import SizesDropdown from "../components/SizesDropdown";
import PopoverOnHover from "../components/Popover";
import { TwoColumnLayout } from "../components/Layouts";
import { PhoneFormField, TextForm } from "../components/FormFields";
import FormActions from "../components/formActions";
import { Chain } from "../api/types";
import { chainGet } from "../api/chain";
import { registerBasicUser } from "../api/login";
import FormJup from "../util/form-jup";
import { ToastContext } from "../providers/ToastProvider";
import { GinParseErrors } from "../util/gin-errors";

interface Params {
  chainUID: string;
}

interface RegisterUserForm {
  name: string;
  email: string;
  phone: string;
  sizes: string[];
  privacyPolicy: boolean;
  newsletter: boolean;
}

export default function Signup() {
  const history = useHistory();
  const { chainUID } = useParams<Params>();
  const [chain, setChain] = useState<Chain | null>(null);
  const { t } = useTranslation();
  const { addToastError } = useContext(ToastContext);
  const [submitted, setSubmitted] = useState(false);
  const [geocoderResult, setGeocoderResult] = useState({
    result: { place_name: "" },
  });
  const [jsValues, setJsValues] = useState({
    address: "",
    sizes: [] as string[],
  });

  // Get chain id from the URL and save to state
  useEffect(() => {
    (async () => {
      if (chainUID) {
        try {
          const chain = (await chainGet(chainUID)).data;
          setChain(chain);
        } catch (e) {
          console.error(`chain ${chainUID} does not exist`);
        }
      }
    })();
  }, [chainUID]);

  // Gather data from form, validate and send to firebase
  function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const values = FormJup<RegisterUserForm>(e);

    if (values.privacyPolicy !== "on") {
      addToastError(t("required") + " " + t("privacyPolicy"));
      return;
    }

    console.log(values);

    (async () => {
      try {
        await registerBasicUser(
          {
            name: values.name,
            email: values.email,
            phone_number: values.phone,
            address: geocoderResult.result.place_name,
            newsletter: values.newsletter === "on",
            sizes: jsValues.sizes,
          },
          chainUID
        );
        setSubmitted(true);
      } catch (e: any) {
        console.error(`Error creating user: ${JSON.stringify(e)}`);
        e.code === "auth/invalid-phone-number"
          ? addToastError(t("pleaseEnterAValid.phoneNumber"))
          : addToastError(
              GinParseErrors(t, e?.data || `Error: ${JSON.stringify(e)}`)
            );
      }
    })();
  }

  if (submitted) {
    return <Redirect to={"/thankyou"} />;
  } else {
    return (
      <>
        <Helmet>
          <title>The Clothing Loop | Signup user</title>
          <meta name="description" content="Signup user" />
        </Helmet>

        <main className="p-10">
          <TwoColumnLayout img="/images/Join-Loop.jpg">
            <div id="container" className="">
              <h1 className="font-semibold text-3xl text-secondary mb-3">
                {t("join")}
                <span> {chain?.name}</span>
              </h1>

              <form onSubmit={onSubmit}>
                <TextForm label={t("name")} name="name" type="text" required />
                <TextForm
                  label={t("email")}
                  name="email"
                  type="email"
                  required
                />

                <PhoneFormField />

                <div className="max-w-xs mb-6">
                  <div className="form-control w-full mb-4">
                    <label className="label">
                      <span className="label-text">{t("address")}</span>
                    </label>
                    <GeocoderSelector onResult={setGeocoderResult} />
                  </div>
                  <SizesDropdown
                    filteredGenders={chain?.genders || []}
                    selectedSizes={jsValues.sizes}
                    handleChange={(s) =>
                      setJsValues((state) => ({ ...state, sizes: s }))
                    }
                  />
                  <PopoverOnHover
                    message={t("weWouldLikeToKnowThisEquallyRepresented")}
                  />
                  <FormActions />
                </div>

                <div className="mb-4">
                  <button
                    type="button"
                    className="btn btn-secondary btn-outline mr-3"
                    onClick={() => history.goBack()}
                  >
                    {t("back")}
                  </button>
                  <button type="submit" className="btn btn-primary">
                    {t("join")}
                    <span className="feather feather-arrow-right ml-4"></span>
                  </button>
                </div>
              </form>
              <div className="text-sm">
                <p className="text">{t("troublesWithTheSignupContactUs")}</p>
                <a
                  className="link"
                  href="mailto:hello@clothingloop.org?subject=Troubles signing up to The Clothing Loop"
                >
                  hello@clothingloop.org
                </a>
              </div>
            </div>
          </TwoColumnLayout>
        </main>
      </>
    );
  }
}
