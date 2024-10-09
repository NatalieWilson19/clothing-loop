import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useStore } from "@nanostores/react";

import type { Chain } from "../../../api/types";
import { $authUser } from "../../../stores/auth";
import useLocalizePath from "../util/localize_path.hooks";

export default function DeleteModal() {
  const { t, i18n } = useTranslation();
  const authUser = useStore($authUser);

  const [chains, setChains] = useState<Chain[]>([]);
  const [showTextArea, setShowTextArea] = useState(false);

  const reasonsForLeaving = [
    t("moved"),
    t("addressToFar"),
    t("notEnoughItemsILiked"),
    t("tooTimeConsuming"),
    t("doneSwapping"),
    t("didntFitIn"),
  ];

  function showTextHandler(event: React.ChangeEvent<HTMLInputElement>) {
    setShowTextArea(!event.target.checked);
  }

  if (!authUser) return;
  const chainNames = authUser.is_root_admin
    ? undefined
    : (authUser.chains
        .filter((uc) => uc.is_chain_admin)
        .map((uc) => chains.find((c) => c.uid === uc.chain_uid))
        .filter((c) => c && c.total_hosts && !(c.total_hosts > 1))
        .map((c) => c!.name) as string[]);

  if (!(chainNames && chainNames.length))
    return (
      <div className="space-y-2">
        <p>Please select a reason for leaving</p>

        <ul className="list-none">
          {reasonsForLeaving.map((r) => (
            <li key={r} className="flex items-center mb-4">
              <input
                type="checkbox"
                className="checkbox border-black"
                name={r}
                id={r}
              />
              <label className="ml-2">{r}</label>
            </li>
          ))}
          <li key="other" className="flex items-center mb-4">
            <input
              type="checkbox"
              className="checkbox border-black"
              name={"other"}
              onChange={(e) => showTextHandler(e)}
            />
            <label className="ml-2">{t("other")}</label>
          </li>
          <li key="other_text" className="">
            <textarea
              disabled={showTextArea}
              className="bg-grey-light w-full"
            />
          </li>
        </ul>
      </div>
    );
  return (
    <div className="space-y-2">
      <p className="">{t("deleteAccountWithLoops")}</p>
      <ul
        className={`text-sm font-semibold mx-8 ${
          chainNames.length > 1 ? "list-disc" : "list-none text-center"
        }`}
      >
        {chainNames.map((name) => (
          <li key={name}>{name}</li>
        ))}
      </ul>
    </div>
  );
}
