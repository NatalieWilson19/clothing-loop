import { useRef, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { useStore } from "@nanostores/react";

import type { Chain } from "../../../api/types";
import { $authUser } from "../../../stores/auth";
import { ReasonsForLeavingI18nKeys } from "../../../api/enums";
import { addToastError } from "../../../stores/toast";

interface DeleteModalProps {
  onSubmitReasonForLeaving: (selectedReasons: string[], other?: string) => void;
  isOpen: boolean;
  onClose: () => void;
  chainNames: string[];
}

export default function DeleteModal({
  onSubmitReasonForLeaving,
  isOpen,
  onClose,
  chainNames,
}: DeleteModalProps) {
  const { t } = useTranslation();
  const authUser = useStore($authUser);

  const [chains, setChains] = useState<Chain[]>([]);
  const [showOtherTextFieldArea, setShowOtherTextFieldArea] = useState(false);
  const [showMovedOptions, setShowMovedOptions] = useState(false);
  const [showNotEnoughItemsOptions, setShowNotEnoughItemsOptions] =
    useState(false);

  const [selectedReasons, setSelectedReasons] = useState<string[]>([]);
  const checkboxesRef = useRef<Record<string, HTMLInputElement | null>>({});

  const moved = Object.keys(ReasonsForLeavingI18nKeys)[0];
  const notEnoughItemsILiked = Object.keys(ReasonsForLeavingI18nKeys)[1];
  const primaryOptions = Object.keys(ReasonsForLeavingI18nKeys).slice(2, 6);
  const other = Object.keys(ReasonsForLeavingI18nKeys)[6];

  const movedOptions = Object.keys(ReasonsForLeavingI18nKeys).slice(7, 10);
  const notEnoughItemsOptions = Object.keys(ReasonsForLeavingI18nKeys).slice(
    10,
  );
  const [otherExplanation, setOtherExplanation] = useState("");

  function handleCheckboxChange(name: string) {
    const checkbox = checkboxesRef.current[name];
    if (!checkbox) return;

    const isChecked = checkbox.checked;
    const updatedReasons = isChecked
      ? [...selectedReasons, name]
      : selectedReasons.filter((reason) => reason !== name);
    setSelectedReasons(updatedReasons);
  }

  function handleOtherTextChange(
    event: React.ChangeEvent<HTMLTextAreaElement>,
  ) {
    const updatedOther = event.target.value;
    setOtherExplanation(updatedOther);
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (selectedReasons.length == 0) {
      addToastError(t("selectReasonForLeaving"));
    } else {
      onSubmitReasonForLeaving(selectedReasons);
      onClose();
    }
  }

  return (
    <dialog
      tabIndex={-1}
      className="fixed overflow-visible inset-0 z-50 open:flex justify-center items-center p-0 shadow-lg backdrop:bg-white/30"
      open={isOpen}
    >
      <div className="space-y-2">
        {chainNames && chainNames.length ? (
          <div>
            <h5 className="text-lg mx-8 my-8">{t("deleteAccountWithLoops")}</h5>
            <ul
              className={`text-sm font-semibold mx-8 ${
                chainNames.length > 1 ? "list-disc" : "list-none text-center"
              }`}
            >
              {chainNames.map((name) => (
                <li key={name}>{name}</li>
              ))}
            </ul>
            <button
              onClick={onClose}
              key="close"
              type="reset"
              className={"btn btn-sm btn-ghost float-end m-6"}
            >
              {t("close")}
            </button>
          </div>
        ) : (
          <div>
            <h5 className="text-lg mx-6 my-6">{t("selectReasonForLeaving")}</h5>
            <form className="bg-white max-w-screen-sm px-6" onSubmit={onSubmit}>
              <ul className="list-none">
                <li key={moved} className="flex items-center mb-4">
                  <input
                    type="checkbox"
                    className="checkbox border-black"
                    name={moved}
                    ref={(el) => (checkboxesRef.current[moved] = el)}
                    onChange={() => {
                      setShowMovedOptions(!showMovedOptions);
                      handleCheckboxChange(moved);
                    }}
                  />
                  <label className="ml-2">
                    {t(ReasonsForLeavingI18nKeys[moved])}
                  </label>
                </li>

                {showMovedOptions ? (
                  <>
                    {movedOptions.map((r) => (
                      <li key={r} className="flex items-center mb-4 ml-8">
                        <input
                          type="checkbox"
                          className="checkbox border-black"
                          name={r}
                          ref={(el) => (checkboxesRef.current[r] = el)}
                          onChange={() => handleCheckboxChange(r)}
                        />
                        <label className="ml-2">
                          {t(ReasonsForLeavingI18nKeys[r])}
                        </label>
                      </li>
                    ))}
                  </>
                ) : null}

                <li
                  key={notEnoughItemsILiked}
                  className="flex items-center mb-4"
                >
                  <input
                    type="checkbox"
                    className="checkbox border-black"
                    name={notEnoughItemsILiked}
                    ref={(el) =>
                      (checkboxesRef.current[notEnoughItemsILiked] = el)
                    }
                    onChange={() => {
                      setShowNotEnoughItemsOptions(!showNotEnoughItemsOptions);
                      handleCheckboxChange(notEnoughItemsILiked);
                    }}
                  />
                  <label className="ml-2">
                    {t(ReasonsForLeavingI18nKeys[notEnoughItemsILiked])}
                  </label>
                </li>

                {showNotEnoughItemsOptions ? (
                  <>
                    {notEnoughItemsOptions.map((r) => (
                      <li key={r} className="flex items-center mb-4 ml-8">
                        <input
                          type="checkbox"
                          className="checkbox border-black"
                          name={r}
                          ref={(el) => (checkboxesRef.current[r] = el)}
                          onChange={() => handleCheckboxChange(r)}
                        />
                        <label className="ml-2">
                          {t(ReasonsForLeavingI18nKeys[r])}
                        </label>
                      </li>
                    ))}
                  </>
                ) : null}
                {primaryOptions.map((r: string) => (
                  <li key={r} className="flex items-center mb-4">
                    <input
                      type="checkbox"
                      className="checkbox border-black"
                      name={r}
                      ref={(el) => (checkboxesRef.current[r] = el)}
                      onChange={() => handleCheckboxChange(r)}
                    />
                    <label className="ml-2">
                      {t(ReasonsForLeavingI18nKeys[r])}
                    </label>
                  </li>
                ))}
                <li key={other} className="flex items-center mb-4">
                  <input
                    type="checkbox"
                    className="checkbox border-black"
                    name={other}
                    ref={(el) => (checkboxesRef.current[other] = el)}
                    onChange={() => {
                      setShowOtherTextFieldArea(!showOtherTextFieldArea);
                      handleCheckboxChange(other);
                    }}
                  />
                  <label className="ml-2">
                    {t(ReasonsForLeavingI18nKeys[other])}
                  </label>
                </li>
                {showOtherTextFieldArea ? (
                  <>
                    <li key="other_textarea" className="mx-7">
                      <label className="text-sm">{t("leaveFeedback")}</label>
                      <textarea
                        required
                        className="bg-grey-light w-full"
                        value={otherExplanation}
                        minLength={5}
                        onChange={handleOtherTextChange}
                      />
                    </li>
                  </>
                ) : null}
              </ul>
              <div className="flex justify-between my-6">
                <button
                  key={"submit"}
                  className="btn btn-sm btn-error"
                  onClick={onSubmit}
                >
                  {t("delete")}
                </button>
                <button
                  onClick={onClose}
                  key="close"
                  type="reset"
                  className={"btn btn-sm btn-ghost"}
                >
                  {t("cancel")}
                </button>
              </div>
            </form>
          </div>
        )}
      </div>
    </dialog>
  );
}
